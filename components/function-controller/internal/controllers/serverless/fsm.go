package serverless

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type stateFn func(context.Context, *reconciler, *systemState) (stateFn, error)

type k8s struct {
	client         resource.Client
	recorder       record.EventRecorder
	statsCollector StatsCollector
}

type cfg struct {
	docker DockerConfig
	fn     FunctionConfig
}

// nolint
type out struct {
	result ctrl.Result
}

type reconciler struct {
	cfg       cfg
	fn        stateFn
	log       *zap.SugaredLogger
	gitClient GitClient
	k8s
	out
}

const (
	continuousGitCheckoutAnnotation = "serverless.kyma-project.io/continuousGitCheckout"
)

func (m *reconciler) reconcile(ctx context.Context, f serverlessv1alpha2.Function) (ctrl.Result, error) {
	state := systemState{instance: f}
	var err error
loop:
	for m.fn != nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break loop

		default:
			m.log.With("stateFn", m.stateFnName()).Info("next state")
			m.fn, err = m.fn(ctx, m, &state)
		}
	}

	m.log.With("requeueAfter", m.result.RequeueAfter).
		With("requeue", m.result.Requeue).
		With("error", err).
		Info("reconciliation result")

	return m.result, err
}

func (m *reconciler) stateFnName() string {
	fullName := runtime.FuncForPC(reflect.ValueOf(m.fn).Pointer()).Name()
	splitFullName := strings.Split(fullName, ".")

	if len(splitFullName) < 3 {
		return fullName
	}
	shortName := splitFullName[2]
	return shortName
}

var (
	// ErrBuildReconcilerFailed is returned if it is impossible to create reconciler build function
	ErrBuildReconcilerFailed = errors.New("build reconciler failed")
)

// this function is a terminator
func buildGenericStatusUpdateStateFn(condition serverlessv1alpha2.Condition, repo *serverlessv1alpha2.Repository, commit string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
		if condition.LastTransitionTime.IsZero() {
			return nil, fmt.Errorf("LastTransitionTime for condition %s is not set", condition.Type)
		}
		existingFunction := &serverlessv1alpha2.Function{}

		err := r.client.Get(ctx, types.NamespacedName{Namespace: s.instance.Namespace, Name: s.instance.Name}, existingFunction)
		if err != nil {
			return nil, client.IgnoreNotFound(err)
		}

		updatedStatus := existingFunction.Status.DeepCopy()
		updatedStatus.Conditions = updateCondition(existingFunction.Status.Conditions, condition)

		if err := r.populateStatusFromSystemState(updatedStatus, s); err != nil {
			return nil, err
		}

		isGitType := s.instance.TypeOf(serverlessv1alpha2.FunctionTypeGit)
		if isGitType && repo != nil {
			updatedStatus.Repository = *repo
			updatedStatus.Commit = commit
		}

		if err := r.updateFunctionStatusWithEvent(ctx, existingFunction, updatedStatus, condition); err != nil {
			r.log.Warnf("while updating function status: %s", err)
			return nil, err
		}
		r.statsCollector.UpdateReconcileStats(&s.instance, condition)
		return nil, nil
	}
}

func (m *reconciler) populateStatusFromSystemState(status *serverlessv1alpha2.FunctionStatus, s *systemState) error {
	status.Runtime = s.instance.Spec.Runtime
	status.RuntimeImageOverride = s.instance.Spec.RuntimeImageOverride

	// set scale sub-resource
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: s.podLabels()})
	if err != nil {
		m.log.Warnf("failed to get selector for labelSelector: %w", err)
		return err
	}
	status.PodSelector = selector.String()

	if len(s.deployments.Items) > 0 {
		status.Replicas = s.deployments.Items[0].Status.Replicas
	}
	return nil
}

func (m *reconciler) updateFunctionStatusWithEvent(ctx context.Context, f *serverlessv1alpha2.Function, s *serverlessv1alpha2.FunctionStatus, condition serverlessv1alpha2.Condition) error {

	if reflect.DeepEqual(f.Status, s) {
		return nil
	}
	f.Status = *s
	if err := m.client.Status().Update(ctx, f); err != nil {
		return err
	}
	eventType := "Normal"
	if condition.Status == corev1.ConditionFalse {
		eventType = "Warning"
	}

	m.recorder.Event(f, eventType, string(condition.Reason), condition.Message)
	return nil
}

func buildStatusUpdateStateFnWithCondition(condition serverlessv1alpha2.Condition) stateFn {
	return buildGenericStatusUpdateStateFn(condition, nil, "")
}

func stateFnGitCheckSources(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	var auth *git.AuthOptions
	if s.instance.Spec.Source.GitRepository.Auth != nil {
		var secret corev1.Secret
		key := client.ObjectKey{
			Namespace: s.instance.Namespace,
			Name:      s.instance.Spec.Source.GitRepository.Auth.SecretName,
		}

		if err := r.client.Get(ctx, key, &secret); err != nil {
			return nil, err
		}

		auth = &git.AuthOptions{
			Type:        git.RepositoryAuthType(s.instance.Spec.Source.GitRepository.Auth.Type),
			Credentials: readSecretData(secret.Data),
			SecretName:  s.instance.Spec.Source.GitRepository.Auth.SecretName,
		}
	}

	options := git.Options{
		URL:       s.instance.Spec.Source.GitRepository.URL,
		Reference: s.instance.Spec.Source.GitRepository.Reference,
		Auth:      auth,
	}

	if skipGitSourceCheck(s.instance, r.cfg) {
		r.log.Info(fmt.Sprintf("skipping function [%s] source check", s.instance.Name))
		expectedJob := s.buildGitJob(options, r.cfg)
		return buildStateFnCheckImageJob(expectedJob), nil
	}

	var revision string
	var err error
	revision, err = r.gitClient.LastCommit(options)
	if err != nil {
		r.log.Error(err, " while fetching last commit")
		var errMsg string
		r.result, errMsg = NextRequeue(err)
		// TODO: This return masks the error from r.syncRevision() and doesn't pass it to the controller. This should be fixed in a follow up PR.
		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonSourceUpdateFailed,
			Message:            errMsg,
		}
		return buildStatusUpdateStateFnWithCondition(condition), nil
	}

	srcChanged := s.gitFnSrcChanged(revision)
	if !srcChanged {
		expectedJob := s.buildGitJob(options, r.cfg)
		return buildStateFnCheckImageJob(expectedJob), nil
	}

	condition := serverlessv1alpha2.Condition{
		Type:               serverlessv1alpha2.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha2.ConditionReasonSourceUpdated,
		Message:            fmt.Sprintf("Sources %s updated", s.instance.Name),
	}

	repository := serverlessv1alpha2.Repository{
		Reference: s.instance.Spec.Source.GitRepository.Reference,
		BaseDir:   s.instance.Spec.Source.GitRepository.BaseDir,
	}

	return buildGenericStatusUpdateStateFn(condition, &repository, revision), nil
}

func stateFnInitialize(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	isGitType := s.instance.TypeOf(serverlessv1alpha2.FunctionTypeGit)
	if isGitType {
		return stateFnGitCheckSources, nil
	}

	return stateFnInlineCheckSources, nil
}

func skipGitSourceCheck(f serverlessv1alpha2.Function, cfg cfg) bool {
	if v, ok := f.Annotations[continuousGitCheckoutAnnotation]; ok && strings.ToLower(v) == "true" {
		return false
	}

	// ConditionConfigurationReady is set to true for git functions when the source is updated.
	// if not, this is a new function, we need to do git check.
	configured := f.Status.Condition(serverlessv1alpha2.ConditionConfigurationReady)
	if configured == nil || !configured.IsTrue() {
		return false
	}

	return time.Since(configured.LastTransitionTime.Time) < cfg.fn.FunctionReadyRequeueDuration
}

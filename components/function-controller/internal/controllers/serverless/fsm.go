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

type stateFn func(context.Context, *reconciler, *systemState) stateFn

type k8s struct {
	client         resource.Client
	recorder       record.EventRecorder
	statsCollector StatsCollector
}

type cfg struct {
	docker DockerConfig
	fn     FunctionConfig
}

//nolint
type out struct {
	err    error
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

loop:
	for m.fn != nil {
		select {
		case <-ctx.Done():
			m.err = ctx.Err()
			break loop
		default:
			m.log.With("stateFn", m.stateFnName()).Info("next state")

			m.fn = m.fn(ctx, m, &state)

		}
	}

	m.log.With("requeueAfter", m.result.RequeueAfter).
		With("requeue", m.result.Requeue).
		With("err", m.err).
		Info("reconciliation result")

	return m.result, m.err
}

func (m *reconciler) stateFnName() string {
	fullName := runtime.FuncForPC(reflect.ValueOf(m.fn).Pointer()).Name()
	splitFullName := strings.Split(fullName, ".")
	shortName := splitFullName[len(splitFullName)-1]
	return shortName
}

var (
	// ErrBuildReconcilerFailed is returned if it is impossible to create reconciler build function
	ErrBuildReconcilerFailed = errors.New("build reconciler failed")
)

// TODO create issue to refactor this
// this function is a terminator
func buildGenericStatusUpdateStateFn(condition serverlessv1alpha2.Condition, repo *serverlessv1alpha2.Repository, commit string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		if condition.LastTransitionTime.IsZero() {
			r.err = fmt.Errorf("LastTransitionTime for condition %s is not set", condition.Type)
			return nil
		}
		currentFunction := &serverlessv1alpha2.Function{}

		r.err = r.client.Get(ctx, types.NamespacedName{Namespace: s.instance.Namespace, Name: s.instance.Name}, currentFunction)
		if r.err != nil {
			r.err = client.IgnoreNotFound(r.err)
			return nil
		}

		currentFunction.Status.Conditions = updateCondition(currentFunction.Status.Conditions, condition)
		// equalConditions := equalConditions(s.instance.Status.Conditions, currentFunction.Status.Conditions)
		equalConditions := reflect.DeepEqual(s.instance.Status.Conditions, currentFunction.Status.Conditions)

		isGitType := s.instance.TypeOf(serverlessv1alpha2.FunctionTypeGit)
		if equalConditions {

			if !isGitType {
				return nil
			}
			// checking if status changed in gitops flow
			if equalRepositories(s.instance.Status.Repository, repo) &&
				s.instance.Status.Commit == commit {
				return nil
			}
		}

		if repo != nil {
			currentFunction.Status.Repository = *repo
			currentFunction.Status.Commit = commit
		}

		currentFunction.Status.Runtime = s.instance.Spec.Runtime
		currentFunction.Status.RuntimeImageOverride = s.instance.Spec.RuntimeImageOverride

		// set scale sub-resource
		selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: s.podLabels()})
		if err != nil {
			r.log.Warnf("failed to get selector for labelSelector: %w", err)
			return nil
		}
		currentFunction.Status.PodSelector = selector.String()

		if len(s.deployments.Items) > 0 {
			currentFunction.Status.Replicas = s.deployments.Items[0].Status.Replicas
		}

		if !equalFunctionStatus(currentFunction.Status, s.instance.Status) {
			if err := r.updateFunctionStatusWithEvent(ctx, currentFunction, condition); err != nil {
				r.log.Warnf("while updating function status: %s", err)
			}
			r.statsCollector.UpdateReconcileStats(&s.instance, condition)
		}
		return nil
	}
}

func (m *reconciler) updateFunctionStatusWithEvent(ctx context.Context, f *serverlessv1alpha2.Function, condition serverlessv1alpha2.Condition) error {
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

// func updateGitFunctionStatus(f *serverlessv1alpha2.Function)

func buildStatusUpdateStateFnWithCondition(condition serverlessv1alpha2.Condition) stateFn {
	return buildGenericStatusUpdateStateFn(condition, nil, "")
}

func stateFnGitCheckSources(ctx context.Context, r *reconciler, s *systemState) stateFn {
	var auth *git.AuthOptions
	if s.instance.Spec.Source.GitRepository.Auth != nil {
		var secret corev1.Secret
		key := client.ObjectKey{
			Namespace: s.instance.Namespace,
			Name:      s.instance.Spec.Source.GitRepository.Auth.SecretName,
		}

		if r.err = r.client.Get(ctx, key, &secret); r.err != nil {
			return nil
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
		return buildStateFnCheckImageJob(expectedJob)
	}

	var revision string
	revision, r.err = r.gitClient.LastCommit(options)
	if r.err != nil {
		r.log.Error(r.err, " while fetching last commit")
		var errMsg string
		r.result, errMsg = NextRequeue(r.err)
		// TODO: This return masks the error from r.syncRevision() and doesn't pass it to the controller. This should be fixed in a follow up PR.
		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonSourceUpdateFailed,
			Message:            errMsg,
		}
		return buildStatusUpdateStateFnWithCondition(condition)
	}

	srcChanged := s.gitFnSrcChanged(revision)
	if !srcChanged {
		expectedJob := s.buildGitJob(options, r.cfg)
		return buildStateFnCheckImageJob(expectedJob)
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

	return buildGenericStatusUpdateStateFn(condition, &repository, revision)
}

func stateFnInitialize(ctx context.Context, r *reconciler, s *systemState) stateFn {
	if r.err = ctx.Err(); r.err != nil {
		return nil
	}

	isGitType := s.instance.TypeOf(serverlessv1alpha2.FunctionTypeGit)
	if isGitType {
		return stateFnGitCheckSources
	}

	return stateFnInlineCheckSources
}

func skipGitSourceCheck(f serverlessv1alpha2.Function, cfg cfg) bool {
	if v, ok := f.Annotations[continuousGitCheckoutAnnotation]; ok && strings.ToLower(v) == "true" {
		return false
	}

	// ConditionConfigurationReady is set to true for git functions when the source is updated. if not, this is a new function, we need to do git check.
	configured := f.Status.Condition(serverlessv1alpha2.ConditionConfigurationReady)
	if configured == nil || !configured.IsTrue() {
		return false
	}

	return time.Since(configured.LastTransitionTime.Time) < cfg.fn.FunctionReadyRequeueDuration
}

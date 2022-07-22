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
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
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
	cfg      cfg
	fn       stateFn
	log      *zap.SugaredLogger
	operator GitOperator
	k8s
	out
}

const (
	continuousGitCheckoutAnnotation = "serverless.kyma-project.io/continuousGitCheckout"
)

func (m *reconciler) reconcile(ctx context.Context, f serverlessv1alpha1.Function) (ctrl.Result, error) {
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
func buildStateFnGenericUpdateStatus(condition serverlessv1alpha1.Condition, repo *serverlessv1alpha1.Repository, commit string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) stateFn {

		condition.LastTransitionTime = metav1.Now()
		currentFunction := &serverlessv1alpha1.Function{}

		r.err = r.client.Get(ctx, types.NamespacedName{Namespace: s.instance.Namespace, Name: s.instance.Name}, currentFunction)
		if r.err != nil {
			r.err = client.IgnoreNotFound(r.err)
			return nil
		}

		currentFunction.Status.Conditions = updateCondition(currentFunction.Status.Conditions, condition)
		equalConditions := equalConditions(s.instance.Status.Conditions, currentFunction.Status.Conditions)

		if equalConditions {

			if s.instance.Spec.Type != serverlessv1alpha1.SourceTypeGit {
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

		currentFunction.Status.Source = s.instance.Spec.Source
		currentFunction.Status.Runtime = serverlessv1alpha1.RuntimeExtended(s.instance.Spec.Runtime)
		currentFunction.Status.RuntimeImageOverride = s.instance.Spec.RuntimeImageOverride

		if !equalFunctionStatus(currentFunction.Status, s.instance.Status) {

			if err := r.client.Status().Update(ctx, currentFunction); err != nil {
				r.log.Warnf("while updating function status: %s", err)
			}

			r.statsCollector.UpdateReconcileStats(&s.instance, condition)

			eventType := "Normal"
			if condition.Status == corev1.ConditionFalse {
				eventType = "Warning"
			}

			r.recorder.Event(currentFunction, eventType, string(condition.Reason), condition.Message)
		}
		return nil
	}
}

func buildStateFnUpdateStateFnFunctionCondition(cdt serverlessv1alpha1.Condition) stateFn {
	return buildStateFnGenericUpdateStatus(cdt, nil, "")
}

func stateFnGitCheckSources(ctx context.Context, r *reconciler, s *systemState) stateFn {
	// read git options
	objectKeyNamespace := s.instance.GetNamespace()
	objectKeyName := s.instance.Spec.Source

	var gitRepository serverlessv1alpha1.GitRepository
	key := client.ObjectKey{
		Namespace: objectKeyNamespace,
		Name:      objectKeyName,
	}

	if r.err = r.client.Get(ctx, key, &gitRepository); r.err != nil {
		return nil
	}

	var auth *git.AuthOptions
	if gitRepository.Spec.Auth != nil {
		var secret corev1.Secret
		key := client.ObjectKey{
			Namespace: s.instance.Namespace,
			Name:      gitRepository.Spec.Auth.SecretName,
		}

		if r.err = r.client.Get(ctx, key, &secret); r.err != nil {
			return nil
		}

		auth = &git.AuthOptions{
			Type:        git.RepositoryAuthType(gitRepository.Spec.Auth.Type),
			Credentials: readSecretData(secret.Data),
			SecretName:  gitRepository.Spec.Auth.SecretName,
		}
	}

	options := git.Options{
		URL:       gitRepository.Spec.URL,
		Reference: s.instance.Spec.Reference,
		Auth:      auth,
	}

	if skipGitSourceCheck(s.instance, r.cfg) {
		r.log.Info(fmt.Sprintf("skipping function [%s] source check", s.instance.Name))
		expectedJob := s.buildGitJob(options, r.cfg)
		return buildStateFnCheckImageJob(expectedJob)
	}

	var revision string
	revision, r.err = r.operator.LastCommit(options)
	if r.err != nil {
		r.log.Error(r.err, "while fetching last commit")
		var errMsg string
		r.result, errMsg = NextRequeue(r.err)
		// TODO: This return masks the error from r.syncRevision() and doesn't pass it to the controller. This should be fixed in a follow up PR.
		condition := serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonSourceUpdateFailed,
			Message:            errMsg,
		}
		return buildStateFnUpdateStateFnFunctionCondition(condition)
	}

	srcChanged := s.gitFnSrcChanged(revision)
	if !srcChanged {
		expectedJob := s.buildGitJob(options, r.cfg)
		return buildStateFnCheckImageJob(expectedJob)
	}

	condition := serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonSourceUpdated,
		Message:            fmt.Sprintf("Sources %s updated", s.instance.Name),
	}

	repository := serverlessv1alpha1.Repository{
		Reference: s.instance.Spec.Reference,
		BaseDir:   s.instance.Spec.BaseDir,
	}

	return buildStateFnGenericUpdateStatus(condition, &repository, revision)
}

func stateFnInitialize(ctx context.Context, r *reconciler, s *systemState) stateFn {
	if r.err = ctx.Err(); r.err != nil {
		return nil
	}

	if s.instance.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		return stateFnGitCheckSources
	}

	return stateFnInlineCheckSources
}

func skipGitSourceCheck(f serverlessv1alpha1.Function, cfg cfg) bool {
	if v, ok := f.Annotations[continuousGitCheckoutAnnotation]; ok && strings.ToLower(v) == "true" {
		return false
	}

	// ConditionConfigurationReady is set to true for git functions when the source is updated. if not, this is a new function, we need to do git check.
	configured := f.Status.Condition(serverlessv1alpha1.ConditionConfigurationReady)
	if configured == nil || !configured.IsTrue() {
		return false
	}

	return time.Since(configured.LastTransitionTime.Time) < cfg.fn.FunctionReadyRequeueDuration
}

package serverless

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
)

var fcManagedByLabel = map[string]string{serverlessv1alpha2.FunctionManagedByLabel: serverlessv1alpha2.FunctionControllerValue}

var backoffLimitExceeded = func(reason string) bool {
	return reason == "BackoffLimitExceeded"
}

// build state function that will check if a job responsible for building function fnImage succeeded or failed;
// if a job is not running start one
func buildStateFnCheckImageJob(expectedJob batchv1.Job) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
		labels := internalFunctionLabels(s.instance)

		err := r.client.ListByLabel(ctx, s.instance.GetNamespace(), labels, &s.jobs)
		if err != nil {
			return nil, errors.Wrap(err, "while listing jobs")
		}

		jobLen := len(s.jobs.Items)

		if jobLen == 0 {
			return buildStateFnRunJob(expectedJob), nil
		}

		jobFailed := s.jobFailed(backoffLimitExceeded)

		conditionStatus := getConditionStatus(
			s.instance.Status.Conditions,
			serverlessv1alpha2.ConditionBuildReady,
		)

		if jobFailed && conditionStatus == corev1.ConditionFalse {
			return stateFnInlineDeleteJobs, nil
		}

		if jobFailed {
			r.result = ctrl.Result{
				RequeueAfter: time.Minute * 5,
				Requeue:      true,
			}

			condition := serverlessv1alpha2.Condition{
				Type:               serverlessv1alpha2.ConditionBuildReady,
				Status:             corev1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             serverlessv1alpha2.ConditionReasonJobFailed,
				Message:            fmt.Sprintf("Job %s failed, it will be re-run", s.jobs.Items[0].Name),
			}
			return buildStatusUpdateStateFnWithCondition(condition), nil
		}

		s.fnImage = s.buildImageAddress(r.cfg.docker.PullAddress)

		diffRuntimeImage, err := functionRuntimeChanged(ctx, r, s)
		if err != nil {
			return nil, errors.Wrap(err, "while checking runtime image change")
		}

		if diffRuntimeImage {
			return stateFnInlineDeleteJobs, nil
		}

		jobChanged := s.fnJobChanged(expectedJob)
		if !jobChanged {
			return stateFnCheckDeployments, nil
		}

		if jobLen > 1 || !equalJobs(s.jobs.Items[0], expectedJob) {
			return stateFnInlineDeleteJobs, nil
		}

		expectedLabels := expectedJob.GetLabels()

		if !mapsEqual(s.jobs.Items[0].GetLabels(), expectedLabels) {
			return buildStateFnInlineUpdateJobLabels(expectedLabels), nil
		}

		return stateFnUpdateJobStatus, nil
	}
}

func functionRuntimeChanged(ctx context.Context, r *reconciler, s *systemState) (bool, error) {
	functionRuntimeImage := s.instance.Status.RuntimeImage
	if functionRuntimeImage == "" {
		return false, nil
	}
	if s.instance.Spec.RuntimeImageOverride != "" {
		result := functionRuntimeImage == s.instance.Spec.RuntimeImageOverride
		return !result, nil
	}

	latestRuntimeImage, err := getRuntimeImageFromConfigMap(ctx, r, s)
	if err != nil {
		return false, errors.Wrap(err, "while fetching runtime image from config map")
	}
	result := latestRuntimeImage == functionRuntimeImage
	return !result, nil
}

func buildStateFnRunJob(expectedJob batchv1.Job) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
		// validate if the max number of running jobs
		// didn't exceed max simultaneous jobs number

		var allJobs batchv1.JobList

		err := r.client.ListByLabel(ctx, "", fcManagedByLabel, &allJobs)
		if err != nil {
			return nil, errors.Wrap(err, "while listing jobs")
		}

		activeJobsCount := countJobs(allJobs, didNotFail, didNotSucceed)
		if activeJobsCount >= r.cfg.fn.Build.MaxSimultaneousJobs {
			r.result = ctrl.Result{
				RequeueAfter: time.Second * 5,
			}
			return nil, nil
		}

		err = r.client.CreateWithReference(ctx, &s.instance, &expectedJob)
		if err != nil {
			return nil, errors.Wrap(err, "while creating job")
		}

		runtimeImage, err := getRuntimeImageFromConfigMap(ctx, r, s)
		if err != nil {
			return nil, errors.Wrap(err, "while extracting runtime fn-image from config map")
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionBuildReady,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonJobCreated,
			Message:            fmt.Sprintf("Job %s created", expectedJob.GetName()),
		}

		s.instance.Status.RuntimeImage = runtimeImage
		return buildStatusUpdateStateFnWithCondition(condition), nil
	}
}

func getRuntimeImageFromConfigMap(ctx context.Context, r *reconciler, s *systemState) (string, error) {
	instance := &corev1.ConfigMap{}
	dockerfileConfigMapName := fmt.Sprintf("dockerfile-%s", s.instance.Status.Runtime)
	err := r.client.Get(ctx, types.NamespacedName{Namespace: s.instance.Namespace, Name: dockerfileConfigMapName}, instance)
	if err != nil {
		return "", errors.Wrap(err, "while extracting correct config map for given runtime")
	}
	baseImage := instance.Data["Dockerfile"]
	re := regexp.MustCompile(`base_image=.*`)
	matchedLines := re.FindStringSubmatch(baseImage)
	if len(matchedLines) == 0 {
		return "", errors.Errorf("could not find the base image from %s", dockerfileConfigMapName)
	}
	runtimeImage := strings.TrimPrefix(matchedLines[0], "base_image=")
	return runtimeImage, err
}

func stateFnInlineDeleteJobs(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	r.log.Info("delete Jobs")

	labels := internalFunctionLabels(s.instance)
	selector := apilabels.SelectorFromSet(labels)

	err := r.client.DeleteAllBySelector(ctx, &batchv1.Job{}, s.instance.GetNamespace(), selector)
	if err != nil {
		return nil, errors.Wrap(err, "while deleting jobs")
	}

	condition := serverlessv1alpha2.Condition{
		Type:               serverlessv1alpha2.ConditionBuildReady,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha2.ConditionReasonJobsDeleted,
		Message:            "Old Jobs deleted",
	}

	return buildStatusUpdateStateFnWithCondition(condition), nil
}

func buildStateFnInlineUpdateJobLabels(m map[string]string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
		s.jobs.Items[0].Labels = m

		jobName := s.jobs.Items[0].GetName()

		r.log.Info(fmt.Sprintf("updating Job %q labels", jobName))

		err := r.client.Update(ctx, &s.jobs.Items[0])
		if err != nil {
			return nil, errors.Wrap(err, "while updating job")
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionBuildReady,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonJobUpdated,
			Message:            fmt.Sprintf("Job %s updated", jobName),
		}

		return buildStatusUpdateStateFnWithCondition(condition), nil
	}
}

func stateFnUpdateJobStatus(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context error")
	}

	job := &s.jobs.Items[0]
	jobName := job.GetName()

	if job.Status.CompletionTime != nil {
		r.log.Info(fmt.Sprintf("job finished %q", jobName))
		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionBuildReady,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonJobFinished,
			Message:            fmt.Sprintf("Job %s finished", jobName),
		}
		return buildStatusUpdateStateFnWithCondition(condition), nil
	}

	if job.Status.Failed < 1 {
		r.log.Info(fmt.Sprintf("job in progress %q", jobName))
		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionBuildReady,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonJobRunning,
			Message:            fmt.Sprintf("Job %s is still in progress", jobName),
		}
		return buildStatusUpdateStateFnWithCondition(condition), nil
	}

	return nil, nil
}

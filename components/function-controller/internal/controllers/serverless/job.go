package serverless

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

var fcManagedByLabel = map[string]string{serverlessv1alpha2.FunctionManagedByLabel: serverlessv1alpha2.FunctionControllerValue}

var backoffLimitExceeded = func(reason string) bool {
	return reason == "BackoffLimitExceeded"
}

// build state function that will check if a job responsible for building function image succeeded or failed;
// if a job is not running start one
func buildStateFnCheckImageJob(expectedJob batchv1.Job) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		labels := s.internalFunctionLabels()

		r.err = r.client.ListByLabel(ctx, s.instance.GetNamespace(), labels, &s.jobs)
		if r.err != nil {
			return nil
		}

		jobLen := len(s.jobs.Items)

		if jobLen == 0 {
			return buildStateFnInlineCreateJob(expectedJob)
		}

		jobFailed := s.jobFailed(backoffLimitExceeded)

		conditionStatus := getConditionStatus(
			s.instance.Status.Conditions,
			serverlessv1alpha2.ConditionBuildReady,
		)

		if jobFailed && conditionStatus == corev1.ConditionFalse {
			return stateFnInlineDeleteJobs
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
			return buildStatusUpdateStateFnWithCondition(condition)
		}

		s.image = s.buildImageAddress(r.cfg.docker.PullAddress)

		jobChanged := s.fnJobChanged(expectedJob)
		if !jobChanged {
			return stateFnCheckDeployments
		}

		if jobLen > 1 || !equalJobs(s.jobs.Items[0], expectedJob) {
			return stateFnInlineDeleteJobs
		}

		expectedLabels := expectedJob.GetLabels()

		if !mapsEqual(s.jobs.Items[0].GetLabels(), expectedLabels) {
			return buildStateFnInlineUpdateJobLabels(expectedLabels)
		}

		return stateFnUpdateJobStatus
	}
}

func buildStateFnInlineCreateJob(expectedJob batchv1.Job) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		// validate if the max number of running jobs
		// didn't exceed max simultaneous jobs number

		var allJobs batchv1.JobList

		r.err = r.client.ListByLabel(ctx, "", fcManagedByLabel, &allJobs)
		if r.err != nil {
			return nil
		}

		activeJobsCount := countJobs(allJobs, didNotFail, didNotSucceed)
		if activeJobsCount >= r.cfg.fn.Build.MaxSimultaneousJobs {
			r.result = ctrl.Result{
				RequeueAfter: time.Second * 5,
			}
			return nil
		}

		r.err = r.client.CreateWithReference(ctx, &s.instance, &expectedJob)
		if r.err != nil {
			return nil
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionBuildReady,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonJobCreated,
			Message:            fmt.Sprintf("Job %s created", expectedJob.GetName()),
		}

		return buildStatusUpdateStateFnWithCondition(condition)
	}
}

func stateFnInlineDeleteJobs(ctx context.Context, r *reconciler, s *systemState) stateFn {
	r.log.Info("delete Jobs")

	labels := s.internalFunctionLabels()
	selector := apilabels.SelectorFromSet(labels)

	r.err = r.client.DeleteAllBySelector(ctx, &batchv1.Job{}, s.instance.GetNamespace(), selector)
	if r.err != nil {
		return nil
	}

	condition := serverlessv1alpha2.Condition{
		Type:               serverlessv1alpha2.ConditionBuildReady,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha2.ConditionReasonJobsDeleted,
		Message:            "Old Jobs deleted",
	}

	return buildStatusUpdateStateFnWithCondition(condition)
}

func buildStateFnInlineUpdateJobLabels(m map[string]string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		s.jobs.Items[0].Labels = m

		jobName := s.jobs.Items[0].GetName()

		r.log.Info(fmt.Sprintf("updating Job %q labels", jobName))

		r.err = r.client.Update(ctx, &s.jobs.Items[0])
		if r.err != nil {
			return nil
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionBuildReady,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonJobUpdated,
			Message:            fmt.Sprintf("Job %s updated", jobName),
		}

		return buildStatusUpdateStateFnWithCondition(condition)
	}
}

func stateFnUpdateJobStatus(ctx context.Context, r *reconciler, s *systemState) stateFn {
	if r.err = ctx.Err(); r.err != nil {
		return nil
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
		return buildStatusUpdateStateFnWithCondition(condition)
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
		return buildStatusUpdateStateFnWithCondition(condition)
	}

	return nil
}

package serverless

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource"

	"github.com/go-logr/logr"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

var fcManagedByLabel = map[string]string{serverlessv1alpha1.FunctionManagedByLabel: serverlessv1alpha1.FunctionControllerValue}

func changeJob(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function, newJob batchv1.Job, jobs []batchv1.Job) (ctrl.Result, error) {
	jobsLen := len(jobs)

	switch {
	case jobsLen == 0:
		activeJobs, err := getActiveAndStatelessJobs(ctx, su.client, log)
		if err != nil {
			return ctrl.Result{}, err
		}

		if activeJobs >= su.config.Build.MaxSimultaneousJobs {
			return ctrl.Result{
				RequeueAfter: time.Second * 5,
			}, nil
		}

		return ctrl.Result{}, createJob(ctx, su, log, instance, newJob)
	case jobsLen > 1 || !equalJobs(jobs[0], newJob):
		return ctrl.Result{}, deleteJobs(ctx, su, log, instance)
	case !mapsEqual(jobs[0].GetLabels(), newJob.GetLabels()):
		return ctrl.Result{}, updateJobLabels(ctx, su, log, instance, jobs[0], newJob.GetLabels())
	default:
		return updateBuildStatus(ctx, su, log, instance, jobs[0])
	}
}

func onJobChange(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function, rtmCfg runtime.Config, configMapName string, jobs []batchv1.Job, dockerConfig DockerConfig) (ctrl.Result, error) {
	newJob := buildJob(instance, rtmCfg, configMapName, dockerConfig, su.config)
	return changeJob(ctx, su, log, instance, newJob, jobs)
}

func equalJobs(existing batchv1.Job, expected batchv1.Job) bool {
	existingArgs := existing.Spec.Template.Spec.Containers[0].Args
	expectedArgs := expected.Spec.Template.Spec.Containers[0].Args

	// Compare destination argument as it contains image tag
	existingDst := getArg(existingArgs, destinationArg)
	expectedDst := getArg(expectedArgs, destinationArg)

	return existingDst == expectedDst
}

func getArg(args []string, arg string) string {
	for _, item := range args {
		if strings.HasPrefix(item, arg) {
			return item
		}
	}
	return ""
}

func createJob(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function, job batchv1.Job) error {
	log.Info("Creating Job")
	if err := su.client.CreateWithReference(ctx, instance, &job); err != nil {
		log.Error(err, "Cannot create Job")
		return err
	}
	log.Info(fmt.Sprintf("Job %s created", job.GetName()))

	return updateStatusWithoutRepository(ctx, su, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionBuildReady,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonJobCreated,
		Message:            fmt.Sprintf("Job %s created", job.GetName()),
	})
}

func deleteJobs(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function) error {
	log.Info("Deleting all old jobs")
	selector := apilabels.SelectorFromSet(internalFunctionLabels(instance))
	if err := su.client.DeleteAllBySelector(ctx, &batchv1.Job{}, instance.GetNamespace(), selector); err != nil {
		log.Error(err, "Cannot delete old Jobs")
		return err
	}
	log.Info("Old Jobs deleted")

	return updateStatusWithoutRepository(ctx, su, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionBuildReady,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonJobsDeleted,
		Message:            "Old Jobs deleted",
	})
}

func updateBuildStatus(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function, job batchv1.Job) (ctrl.Result, error) {
	switch {
	case job.Status.CompletionTime != nil:
		log.Info(fmt.Sprintf("Job %s finished", job.GetName()))
		return ctrl.Result{}, updateStatusWithoutRepository(ctx, su, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobFinished,
			Message:            fmt.Sprintf("Job %s finished", job.GetName()),
		})
	case job.Status.Failed < 1:
		log.Info(fmt.Sprintf("Job %s is still in progress", job.GetName()))
		return ctrl.Result{}, updateStatusWithoutRepository(ctx, su, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobRunning,
			Message:            fmt.Sprintf("Job %s is still in progress", job.GetName()),
		})
	default:
		log.Info(fmt.Sprintf("Job %s failed", job.GetName()))
		return ctrl.Result{RequeueAfter: su.config.RequeueDuration}, updateStatusWithoutRepository(ctx, su, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobFailed,
			Message:            fmt.Sprintf("Job %s failed", job.GetName()),
		})
	}
}

func updateJobLabels(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function, job batchv1.Job, newLabels map[string]string) error {
	newJob := job.DeepCopy()
	newJob.Labels = newLabels

	log.Info(fmt.Sprintf("Updating labels of a Job with name %s", newJob.GetName()))
	if err := su.client.Update(ctx, newJob); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update Job with name %s", newJob.GetName()))
		return err
	}
	log.Info(fmt.Sprintf("Job %s updated", newJob.GetName()))

	return updateStatusWithoutRepository(ctx, su, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionBuildReady,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonJobUpdated,
		Message:            fmt.Sprintf("Job %s updated", newJob.GetName()),
	})
}

func getActiveAndStatelessJobs(ctx context.Context, client resource.Client, log logr.Logger) (int, error) {
	var allJobs batchv1.JobList
	if err := client.ListByLabel(ctx, "", fcManagedByLabel, &allJobs); err != nil {
		log.Error(err, "Cannot list Jobs")
		return 0, err
	}

	out := 0
	for _, j := range allJobs.Items {
		if j.Status.Succeeded == 0 && j.Status.Failed == 0 {
			out++
		}
	}
	return out, nil
}

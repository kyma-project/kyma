package serverless

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

var fcManagedByLabel = map[string]string{serverlessv1alpha1.FunctionManagedByLabel: "function-controller"}

func (r *FunctionReconciler) isOnJobChange(instance *serverlessv1alpha1.Function, rtmCfg runtime.Config, jobs []batchv1.Job, deployments []appsv1.Deployment, gitOptions git.Options) bool {
	image := r.buildImageAddress(instance)
	buildStatus := r.getConditionStatus(instance.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)

	var expectedJob batchv1.Job
	if instance.Spec.Type != serverlessv1alpha1.SourceTypeGit {
		expectedJob = r.buildJob(instance, rtmCfg, "")
	} else {
		expectedJob = r.buildGitJob(instance, gitOptions, rtmCfg)
	}

	if len(deployments) == 1 &&
		deployments[0].Spec.Template.Spec.Containers[0].Image == image &&
		buildStatus != corev1.ConditionUnknown &&
		len(jobs) > 0 &&
		r.mapsEqual(expectedJob.GetLabels(), jobs[0].GetLabels()) {
		return buildStatus == corev1.ConditionFalse
	}

	return len(jobs) != 1 ||
		len(jobs[0].Spec.Template.Spec.Containers) != 1 ||
		// Compare image argument
		!r.equalJobs(jobs[0], expectedJob) ||
		!r.mapsEqual(expectedJob.GetLabels(), jobs[0].GetLabels()) ||
		buildStatus == corev1.ConditionUnknown ||
		buildStatus == corev1.ConditionFalse
}

func (r *FunctionReconciler) changeJob(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, newJob batchv1.Job, jobs []batchv1.Job) (ctrl.Result, error) {
	jobsLen := len(jobs)

	switch {
	case jobsLen == 0:
		activeJobs, err := r.getActiveAndStatelessJobs(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}

		if activeJobs >= r.config.Build.MaxSimultaneousJobs {
			return ctrl.Result{
				RequeueAfter: time.Second * 5,
			}, nil
		}

		return r.createJob(ctx, log, instance, newJob)
	case jobsLen > 1 || !r.equalJobs(jobs[0], newJob):
		return r.deleteJobs(ctx, log, instance)
	case !r.mapsEqual(jobs[0].GetLabels(), newJob.GetLabels()):
		return r.updateJobLabels(ctx, log, instance, jobs[0], newJob.GetLabels())
	default:
		return r.updateBuildStatus(ctx, log, instance, jobs[0])
	}
}

func (r *FunctionReconciler) onGitJobChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, rtmCfg runtime.Config, jobs []batchv1.Job, gitOptions git.Options) (ctrl.Result, error) {
	newJob := r.buildGitJob(instance, gitOptions, rtmCfg)
	return r.changeJob(ctx, log, instance, newJob, jobs)
}
func (r *FunctionReconciler) onJobChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, rtmCfg runtime.Config, configMapName string, jobs []batchv1.Job) (ctrl.Result, error) {
	newJob := r.buildJob(instance, rtmCfg, configMapName)
	return r.changeJob(ctx, log, instance, newJob, jobs)
}

func (r *FunctionReconciler) equalJobs(existing batchv1.Job, expected batchv1.Job) bool {
	existingArgs := existing.Spec.Template.Spec.Containers[0].Args
	expectedArgs := expected.Spec.Template.Spec.Containers[0].Args

	// Compare destination argument as it contains image tag
	existingDst := r.getArg(existingArgs, destinationArg)
	expectedDst := r.getArg(expectedArgs, destinationArg)

	return existingDst == expectedDst
}

func (r *FunctionReconciler) getArg(args []string, arg string) string {
	for _, item := range args {
		if strings.HasPrefix(item, arg) {
			return item
		}
	}
	return ""
}

func (r *FunctionReconciler) createJob(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, job batchv1.Job) (ctrl.Result, error) {
	log.Info("Creating Job")
	if err := r.client.CreateWithReference(ctx, instance, &job); err != nil {
		log.Error(err, "Cannot create Job")
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Job %s created", job.GetName()))

	return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionBuildReady,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonJobCreated,
		Message:            fmt.Sprintf("Job %s created", job.GetName()),
	})
}

func (r *FunctionReconciler) deleteJobs(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function) (ctrl.Result, error) {
	log.Info("Deleting all old jobs")
	selector := apilabels.SelectorFromSet(r.internalFunctionLabels(instance))
	if err := r.client.DeleteAllBySelector(ctx, &batchv1.Job{}, instance.GetNamespace(), selector); err != nil {
		log.Error(err, "Cannot delete old Jobs")
		return ctrl.Result{}, err
	}
	log.Info("Old Jobs deleted")

	return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionBuildReady,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonJobsDeleted,
		Message:            "Old Jobs deleted",
	})
}

func (r *FunctionReconciler) updateBuildStatus(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, job batchv1.Job) (ctrl.Result, error) {
	switch {
	case job.Status.CompletionTime != nil:
		log.Info(fmt.Sprintf("Job %s finished", job.GetName()))
		return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobFinished,
			Message:            fmt.Sprintf("Job %s finished", job.GetName()),
		})
	case job.Status.Failed < 1:
		log.Info(fmt.Sprintf("Job %s is still in progress", job.GetName()))
		return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobRunning,
			Message:            fmt.Sprintf("Job %s is still in progress", job.GetName()),
		})
	default:
		log.Info(fmt.Sprintf("Job %s failed", job.GetName()))
		return r.updateStatusWithoutRepository(ctx, ctrl.Result{RequeueAfter: r.config.RequeueDuration}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobFailed,
			Message:            fmt.Sprintf("Job %s failed", job.GetName()),
		})
	}
}

func (r *FunctionReconciler) updateJobLabels(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, job batchv1.Job, newLabels map[string]string) (ctrl.Result, error) {
	newJob := job.DeepCopy()
	newJob.Labels = newLabels

	log.Info(fmt.Sprintf("Updating labels of a Job with name %s", newJob.GetName()))
	if err := r.client.Update(ctx, newJob); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update Job with name %s", newJob.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Job %s updated", newJob.GetName()))

	return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionBuildReady,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonJobUpdated,
		Message:            fmt.Sprintf("Job %s updated", newJob.GetName()),
	})
}

func (r *FunctionReconciler) getActiveAndStatelessJobs(ctx context.Context) (int, error) {
	var allJobs batchv1.JobList
	if err := r.client.ListByLabel(ctx, "", fcManagedByLabel, &allJobs); err != nil {
		r.Log.Error(err, "Cannot list Jobs")
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

package serverless

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func (r *FunctionReconciler) isOnJobChange(instance *serverlessv1alpha1.Function, jobs []batchv1.Job, deployments []appsv1.Deployment) bool {
	image := r.buildExternalImageAddress(instance)
	buildStatus := r.getConditionStatus(instance.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)

	if len(deployments) == 1 && deployments[0].Spec.Template.Spec.Containers[0].Image == image && buildStatus != corev1.ConditionUnknown {
		return false
	}

	expectedJob := r.buildJob(instance, "")

	return len(jobs) != 1 ||
		len(jobs[0].Spec.Template.Spec.Containers) != 1 ||
		// Compare image argument
		!r.equalJobs(jobs[0], expectedJob) ||
		buildStatus == corev1.ConditionUnknown ||
		buildStatus == corev1.ConditionFalse
}

func (r *FunctionReconciler) onJobChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, configMapName string, jobs []batchv1.Job) (ctrl.Result, error) {
	newJob := r.buildJob(instance, configMapName)
	jobsLen := len(jobs)

	switch {
	case jobsLen == 0:
		return r.createJob(ctx, log, instance, newJob)
	case jobsLen > 1 || !r.equalJobs(jobs[0], newJob):
		return r.deleteJobs(ctx, log, instance)
	default:
		return r.updateBuildStatus(ctx, log, instance, jobs[0])
	}
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

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionBuildReady,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonJobCreated,
		Message:            fmt.Sprintf("Job %s created", job.GetName()),
	})
}

func (r *FunctionReconciler) deleteJobs(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function) (ctrl.Result, error) {
	log.Info("Deleting all old jobs")
	selector := apilabels.SelectorFromSet(r.functionLabels(instance))
	if err := r.client.DeleteAllBySelector(ctx, &batchv1.Job{}, instance.GetNamespace(), selector); err != nil {
		log.Error(err, "Cannot delete old Jobs")
		return ctrl.Result{}, err
	}
	log.Info("Old Jobs deleted")

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
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
		return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobFinished,
			Message:            fmt.Sprintf("Job %s finished", job.GetName()),
		})
	case job.Status.Failed < 1:
		log.Info(fmt.Sprintf("Job %s is still in progress", job.GetName()))
		return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobRunning,
			Message:            fmt.Sprintf("Job %s is still in progress", job.GetName()),
		})
	default:
		log.Info(fmt.Sprintf("Job %s failed", job.GetName()))
		return r.updateStatus(ctx, ctrl.Result{RequeueAfter: r.config.RequeueDuration}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobFailed,
			Message:            fmt.Sprintf("Job %s failed", job.GetName()),
		})
	}
}

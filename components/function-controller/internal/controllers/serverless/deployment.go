package serverless

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func (r *FunctionReconciler) isOnDeploymentChange(instance *serverlessv1alpha1.Function, deployments []appsv1.Deployment) bool {
	deployStatus := r.getConditionStatus(instance.Status.Conditions, serverlessv1alpha1.ConditionRunning)
	expectedDeployment := r.buildDeployment(instance)

	return !(len(deployments) == 1 &&
		len(deployments[0].Spec.Template.Spec.Containers) == 1 &&
		// Compare image argument
		r.equalDeployments(deployments[0], expectedDeployment) &&
		deployStatus == corev1.ConditionUnknown)
}

func (r *FunctionReconciler) onDeploymentChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, deployments []appsv1.Deployment) (ctrl.Result, error) {
	newDeployment := r.buildDeployment(instance)

	switch {
	case len(deployments) == 0:
		return r.createDeployment(ctx, log, instance, newDeployment)
	case !r.equalDeployments(deployments[0], newDeployment):
		return r.updateDeployment(ctx, log, instance, deployments[0], newDeployment)
	default:
		return r.updateDeploymentStatus(ctx, log, instance, deployments[0])
	}
}

func (r *FunctionReconciler) createDeployment(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, deployment appsv1.Deployment) (ctrl.Result, error) {
	log.Info(fmt.Sprintf("Creating Deployment %s", deployment.GetName()))
	if err := r.client.CreateWithReference(ctx, instance, &deployment); err != nil {
		log.Error(err, fmt.Sprintf("Cannot create Deployment with name %s", deployment.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Deployment %s created", deployment.GetName()))

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonDeploymentCreated,
		Message:            fmt.Sprintf("Deployment %s created", deployment.GetName()),
	})
}

func (r *FunctionReconciler) updateDeployment(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, oldDeployment appsv1.Deployment, newDeployment appsv1.Deployment) (ctrl.Result, error) {
	deploy := oldDeployment.DeepCopy()
	deploy.Spec = newDeployment.Spec
	deploy.ObjectMeta.Labels = newDeployment.GetLabels()

	log.Info(fmt.Sprintf("Updating Deployment %s", deploy.GetName()))
	if err := r.client.Update(ctx, deploy); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update Deployment with name %s", deploy.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Deployment %s updated", deploy.GetName()))

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonDeploymentUpdated,
		Message:            fmt.Sprintf("Deployment %s updated", deploy.GetName()),
	})
}

func (r *FunctionReconciler) equalDeployments(existing appsv1.Deployment, expected appsv1.Deployment) bool {
	return len(existing.Spec.Template.Spec.Containers) > 0 &&
		len(existing.Spec.Template.Spec.Containers) == len(expected.Spec.Template.Spec.Containers) &&
		existing.Spec.Template.Spec.Containers[0].Image == expected.Spec.Template.Spec.Containers[0].Image &&
		r.envsEqual(existing.Spec.Template.Spec.Containers[0].Env, expected.Spec.Template.Spec.Containers[0].Env) &&
		r.mapsEqual(existing.GetLabels(), expected.GetLabels()) &&
		r.mapsEqual(existing.Spec.Template.GetLabels(), expected.Spec.Template.GetLabels()) &&
		r.mapsEqual(existing.Spec.Template.GetAnnotations(), expected.Spec.Template.GetAnnotations()) &&
		equalResources(existing.Spec.Template.Spec.Containers[0].Resources, expected.Spec.Template.Spec.Containers[0].Resources)
}

func (r *FunctionReconciler) updateDeploymentStatus(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, deployment appsv1.Deployment) (ctrl.Result, error) {
	switch {
	case r.isDeploymentReady(deployment):
		log.Info(fmt.Sprintf("Deployment %s is ready", deployment.GetName()))
		return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonDeploymentReady,
			Message:            fmt.Sprintf("Function %s is ready", instance.GetName()),
		})
	case r.isDeploymentInProgress(deployment):
		log.Info(fmt.Sprintf("Deployment %s is not ready yet", deployment.GetName()))
		return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonDeploymentWaiting,
			Message:            fmt.Sprintf("Deployment %s is not ready yet", deployment.GetName()),
		})
	default:
		log.Info(fmt.Sprintf("Deployment %s failed", deployment.GetName()))
		return r.updateStatus(ctx, ctrl.Result{RequeueAfter: r.config.RequeueDuration}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonDeploymentFailed,
			Message:            fmt.Sprintf("Deployment %s failed", deployment.GetName()),
		})
	}
}

func (r *FunctionReconciler) isDeploymentInProgress(deployment appsv1.Deployment) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentProgressing {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}

func (r *FunctionReconciler) isDeploymentReady(deployment appsv1.Deployment) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}

package serverless

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const (
	// Progressing:
	// NewRSAvailableReason is added in a deployment when its newest replica set is made available
	// ie. the number of new pods that have passed readiness checks and run for at least minReadySeconds
	// is at least the minimum available pods that need to run for the deployment.
	NewRSAvailableReason = "NewReplicaSetAvailable"

	// Available:
	// MinimumReplicasAvailable is added in a deployment when it has its minimum replicas required available.
	MinimumReplicasAvailable = "MinimumReplicasAvailable"
)

func (r *FunctionReconciler) isOnDeploymentChange(instance *serverlessv1alpha1.Function, rtmConfig runtime.Config, deployments []appsv1.Deployment, dockerConfig DockerConfig) bool {
	expectedDeployment := r.buildDeployment(instance, rtmConfig, dockerConfig)
	resourceOk := len(deployments) == 1 && r.equalDeployments(deployments[0], expectedDeployment, isScalingEnabled(instance))

	return !resourceOk
}

func (r *FunctionReconciler) onDeploymentChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, rtmConfig runtime.Config, deployments []appsv1.Deployment, dockerConfig DockerConfig) (ctrl.Result, error) {
	newDeployment := r.buildDeployment(instance, rtmConfig, dockerConfig)

	switch {
	case len(deployments) == 0:
		return r.createDeployment(ctx, log, instance, newDeployment)
	case len(deployments) > 1: // this step is needed, as sometimes informers lag behind reality, and then we create 2 (or more) deployments by accident
		return r.deleteAllDeployments(ctx, instance, log)
	case !r.equalDeployments(deployments[0], newDeployment, isScalingEnabled(instance)):
		return r.updateDeployment(ctx, log, instance, deployments[0], newDeployment)
	default:
		return r.updateDeploymentStatus(ctx, log, instance, deployments, corev1.ConditionUnknown)
	}
}

func (r *FunctionReconciler) createDeployment(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, deployment appsv1.Deployment) (ctrl.Result, error) {
	log.Info("Creating Deployment")
	if err := r.client.CreateWithReference(ctx, instance, &deployment); err != nil {
		log.Error(err, fmt.Sprintf("Cannot create Deployment with name %s", deployment.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Deployment %s created", deployment.GetName()))

	return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
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

	return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonDeploymentUpdated,
		Message:            fmt.Sprintf("Deployment %s updated", deploy.GetName()),
	})
}

func (r *FunctionReconciler) equalDeployments(existing appsv1.Deployment, expected appsv1.Deployment, scalingEnabled bool) bool {
	return len(existing.Spec.Template.Spec.Containers) == 1 &&
		len(existing.Spec.Template.Spec.Containers) == len(expected.Spec.Template.Spec.Containers) &&
		existing.Spec.Template.Spec.Containers[0].Image == expected.Spec.Template.Spec.Containers[0].Image &&
		r.envsEqual(existing.Spec.Template.Spec.Containers[0].Env, expected.Spec.Template.Spec.Containers[0].Env) &&
		r.mapsEqual(existing.GetLabels(), expected.GetLabels()) &&
		r.mapsEqual(existing.Spec.Template.GetLabels(), expected.Spec.Template.GetLabels()) &&
		r.mapsEqual(existing.Spec.Template.GetAnnotations(), expected.Spec.Template.GetAnnotations()) &&
		equalResources(existing.Spec.Template.Spec.Containers[0].Resources, expected.Spec.Template.Spec.Containers[0].Resources) &&
		(scalingEnabled || equalInt32Pointer(existing.Spec.Replicas, expected.Spec.Replicas))
}

func (r *FunctionReconciler) updateDeploymentStatus(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, deployments []appsv1.Deployment, runningStatus corev1.ConditionStatus) (ctrl.Result, error) {
	switch {
	// this step is both in onDeploymentChange and as last step in reconcile
	// it's checked here in onDeploymentChange to prevent nasty data races, where somehow deployment becomes ready before we
	// trigger next reconcile loop, in which we should create svc
	case r.isDeploymentReady(deployments[0]):
		log.Info(fmt.Sprintf("Deployment %s is ready", deployments[0].GetName()))
		return r.updateStatusWithoutRepository(ctx, ctrl.Result{
			RequeueAfter: r.config.FunctionReadyRequeueDuration,
		}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             runningStatus,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonDeploymentReady,
			Message:            fmt.Sprintf("Deployment %s is ready", deployments[0].GetName()),
		})
	case r.hasDeploymentConditionTrueStatus(deployments[0], appsv1.DeploymentProgressing):
		log.Info(fmt.Sprintf("Deployment %s is not ready yet", deployments[0].GetName()))
		return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonDeploymentWaiting,
			Message:            fmt.Sprintf("Deployment %s is not ready yet", deployments[0].GetName()),
		})
	default:
		log.Info(fmt.Sprintf("Deployment %s failed", deployments[0].GetName()))
		yamlConditions, err := yaml.Marshal(deployments[0].Status.Conditions)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "while marshalling deployment status to yaml")
		}
		return r.updateStatusWithoutRepository(ctx, ctrl.Result{RequeueAfter: r.config.RequeueDuration}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonDeploymentFailed,
			Message:            fmt.Sprintf("Deployment %s failed with condition: \n%s", deployments[0].GetName(), yamlConditions),
		})
	}
}

func (r *FunctionReconciler) isDeploymentReady(deployment appsv1.Deployment) bool {
	return r.hasDeploymentConditionTrueStatusWithReason(deployment, appsv1.DeploymentAvailable, MinimumReplicasAvailable) &&
		r.hasDeploymentConditionTrueStatusWithReason(deployment, appsv1.DeploymentProgressing, NewRSAvailableReason)
}

func (r *FunctionReconciler) hasDeploymentConditionTrueStatus(deployment appsv1.Deployment, conditionType appsv1.DeploymentConditionType) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func (r *FunctionReconciler) hasDeploymentConditionTrueStatusWithReason(deployment appsv1.Deployment, conditionType appsv1.DeploymentConditionType, reason string) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionTrue &&
				condition.Reason == reason
		}
	}
	return false
}

func equalResources(existing, expected corev1.ResourceRequirements) bool {
	return existing.Requests.Memory().Equal(*expected.Requests.Memory()) &&
		existing.Requests.Cpu().Equal(*expected.Requests.Cpu()) &&
		existing.Limits.Memory().Equal(*expected.Limits.Memory()) &&
		existing.Limits.Cpu().Equal(*expected.Limits.Cpu())
}

func (r *FunctionReconciler) deleteAllDeployments(ctx context.Context, instance *serverlessv1alpha1.Function, log logr.Logger) (ctrl.Result, error) {
	log.Info("Deleting all underlying Deployments")
	selector := apilabels.SelectorFromSet(r.internalFunctionLabels(instance))
	if err := r.client.DeleteAllBySelector(ctx, &appsv1.Deployment{}, instance.GetNamespace(), selector); err != nil {
		log.Error(err, "Cannot delete underlying Deployments")
		return ctrl.Result{}, err
	}

	log.Info("Underlying Deployments deleted")
	return ctrl.Result{}, nil
}

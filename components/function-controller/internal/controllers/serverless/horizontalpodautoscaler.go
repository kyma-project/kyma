package serverless

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func (r *FunctionReconciler) isOnHorizontalPodAutoscalerChange(instance *serverlessv1alpha1.Function, hpas []autoscalingv1.HorizontalPodAutoscaler, deployments []appsv1.Deployment) bool {
	if len(deployments) == 0 {
		return false
	}

	newHpa := r.buildHorizontalPodAutoscaler(instance, deployments[0].GetName())
	return !(len(hpas) == 1 &&
		r.equalHorizontalPodAutoscalers(hpas[0], newHpa))
}

func (r *FunctionReconciler) onHorizontalPodAutoscalerChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, hpas []autoscalingv1.HorizontalPodAutoscaler, deploymentName string) (ctrl.Result, error) {
	newHpa := r.buildHorizontalPodAutoscaler(instance, deploymentName)

	switch {
	case len(hpas) == 0:
		return r.createHorizontalPodAutoscaler(ctx, log, instance, newHpa)
	case len(hpas) > 1: // this step is needed, as sometimes informers lag behind reality, and then we create 2 (or more) hpas by accident
		return r.deleteAllHorizontalPodAutoscalers(ctx, instance, log)
	case !r.equalHorizontalPodAutoscalers(hpas[0], newHpa):
		return r.updateHorizontalPodAutoscaler(ctx, log, instance, hpas[0], newHpa)
	default:
		log.Info(fmt.Sprintf("HorizontalPodAutoscaler %s is ready", hpas[0].GetName()))
		return ctrl.Result{}, nil
	}
}

func equalInt32Pointer(first *int32, second *int32) bool {
	// comment
	if first == nil && second == nil {
		return true
	}
	if (first != nil && second == nil) || (first == nil && second != nil) {
		return false
	}

	return *first == *second
}

func (r *FunctionReconciler) equalHorizontalPodAutoscalers(existing, expected autoscalingv1.HorizontalPodAutoscaler) bool {
	return equalInt32Pointer(existing.Spec.TargetCPUUtilizationPercentage, expected.Spec.TargetCPUUtilizationPercentage) &&
		equalInt32Pointer(existing.Spec.MinReplicas, expected.Spec.MinReplicas) &&
		existing.Spec.MaxReplicas == expected.Spec.MaxReplicas &&
		r.mapsEqual(existing.Labels, expected.Labels) &&
		existing.Spec.ScaleTargetRef.Name == expected.Spec.ScaleTargetRef.Name
}

func (r *FunctionReconciler) createHorizontalPodAutoscaler(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, hpa autoscalingv1.HorizontalPodAutoscaler) (ctrl.Result, error) {
	log.Info("Creating HorizontalPodAutoscaler")
	if err := r.client.CreateWithReference(ctx, instance, &hpa); err != nil {
		log.Error(err, fmt.Sprintf("Cannot create HorizontalPodAutoscaler with name %s", hpa.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("HorizontalPodAutoscaler %s created", hpa.GetName()))

	return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonHorizontalPodAutoscalerCreated,
		Message:            fmt.Sprintf("HorizontalPodAutoscaler %s created", hpa.GetName()),
	})
}

func (r *FunctionReconciler) updateHorizontalPodAutoscaler(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, oldHpa, newHpa autoscalingv1.HorizontalPodAutoscaler) (ctrl.Result, error) {
	hpaCopy := oldHpa.DeepCopy()
	hpaCopy.Spec = newHpa.Spec
	hpaCopy.ObjectMeta.Labels = newHpa.GetLabels()

	log.Info(fmt.Sprintf("Updating HorizontalPodAutoscaler %s", hpaCopy.GetName()))
	if err := r.client.Update(ctx, hpaCopy); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update HorizontalPodAutoscaler with name %s", hpaCopy.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("HorizontalPodAutoscaler %s updated", hpaCopy.GetName()))

	return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonHorizontalPodAutoscalerUpdated,
		Message:            fmt.Sprintf("HorizontalPodAutoscaler %s updated", hpaCopy.GetName()),
	})
}

func (r *FunctionReconciler) deleteAllHorizontalPodAutoscalers(ctx context.Context, instance *serverlessv1alpha1.Function, log logr.Logger) (ctrl.Result, error) {
	log.Info("Deleting all HorizontalPodAutoscalers")
	selector := apilabels.SelectorFromSet(r.internalFunctionLabels(instance))
	if err := r.client.DeleteAllBySelector(ctx, &autoscalingv1.HorizontalPodAutoscaler{}, instance.GetNamespace(), selector); err != nil {
		log.Error(err, "Cannot delete underlying HorizontalPodAutoscalers")
		return ctrl.Result{}, err
	}

	log.Info("Underlying HorizontalPodAutoscalers deleted")
	return ctrl.Result{}, nil
}

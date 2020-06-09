package serverless

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func (r *FunctionReconciler) isOnServiceChange(instance *serverlessv1alpha1.Function, services []corev1.Service) bool {
	newSvc := r.buildService(instance)
	return !(len(services) == 1 &&
		r.equalServices(services[0], newSvc))
}

func (r *FunctionReconciler) onServiceChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, services []corev1.Service) (ctrl.Result, error) {
	newSvc := r.buildService(instance)

	switch {
	case len(services) == 0:
		return r.createService(ctx, log, instance, newSvc)
	case len(services) > 1:
		log.Info("Too many Services for function")
		return r.updateStatus(ctx, ctrl.Result{RequeueAfter: r.config.RequeueDuration}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonServiceError,
			Message:            fmt.Sprintf("Service step failed, too many Services found, needed 1 but got %d", len(services)),
		})
	case !r.equalServices(services[0], newSvc):
		return r.updateService(ctx, log, instance, services[0], newSvc)
	default:
		log.Info(fmt.Sprintf("Service %s is ready", services[0].GetName()))
		return ctrl.Result{}, nil
	}
}

func (r *FunctionReconciler) equalServices(existing corev1.Service, expected corev1.Service) bool {
	return r.mapsEqual(existing.Spec.Selector, expected.Spec.Selector) &&
		r.mapsEqual(existing.Labels, expected.Labels) &&
		len(existing.Spec.Ports) == len(expected.Spec.Ports) &&
		len(expected.Spec.Ports) > 0 &&
		len(existing.Spec.Ports) > 0 &&
		existing.Spec.Ports[0].String() == expected.Spec.Ports[0].String()
}

func (r *FunctionReconciler) createService(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, service corev1.Service) (ctrl.Result, error) {
	log.Info(fmt.Sprintf("Creating Service %s", service.GetName()))
	if err := r.client.CreateWithReference(ctx, instance, &service); err != nil {
		log.Error(err, fmt.Sprintf("Cannot create Service with name %s", service.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Service %s created", service.GetName()))

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonServiceCreated,
		Message:            fmt.Sprintf("Service %s created", service.GetName()),
	})
}

func (r *FunctionReconciler) updateService(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, oldService corev1.Service, newService corev1.Service) (ctrl.Result, error) {
	svc := oldService.DeepCopy()

	// manually change fields that interest us, as clusterIP is immutable
	svc.Spec.Ports = newService.Spec.Ports
	svc.Spec.Selector = newService.Spec.Selector
	svc.Spec.Type = newService.Spec.Type

	svc.ObjectMeta.Labels = newService.GetLabels()

	log.Info(fmt.Sprintf("Updating Service %s", svc.GetName()))
	if err := r.client.Update(ctx, svc); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update Service with name %s", svc.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Service %s updated", svc.GetName()))

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonServiceUpdated,
		Message:            fmt.Sprintf("Service %s updated", svc.GetName()),
	})
}

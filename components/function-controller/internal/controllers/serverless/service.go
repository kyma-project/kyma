package serverless

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func isOnServiceChange(instance *serverlessv1alpha1.Function, services []corev1.Service) bool {
	newSvc := buildService(instance)
	return !(len(services) == 1 &&
		equalServices(services[0], newSvc))
}

func onServiceChange(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function, services []corev1.Service) error {
	newSvc := buildService(instance)

	switch {
	case len(services) == 0:
		return createService(ctx, su, log, instance, newSvc)
	case len(services) > 1:
		return deleteExcessServices(ctx, su.client, instance, log, services)
	case !equalServices(services[0], newSvc):
		return updateService(ctx, su, log, instance, services[0], newSvc)
	default:
		log.Info(fmt.Sprintf("Service %s is ready", services[0].GetName()))
		return nil
	}
}

func equalServices(existing corev1.Service, expected corev1.Service) bool {
	return mapsEqual(existing.Spec.Selector, expected.Spec.Selector) &&
		mapsEqual(existing.Labels, expected.Labels) &&
		len(existing.Spec.Ports) == len(expected.Spec.Ports) &&
		len(expected.Spec.Ports) > 0 &&
		len(existing.Spec.Ports) > 0 &&
		existing.Spec.Ports[0].String() == expected.Spec.Ports[0].String()
}

func createService(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function, service corev1.Service) error {
	log.Info(fmt.Sprintf("Creating Service %s", service.GetName()))
	if err := su.client.CreateWithReference(ctx, instance, &service); err != nil {
		log.Error(err, fmt.Sprintf("Cannot create Service with name %s", service.GetName()))
		return err
	}
	log.Info(fmt.Sprintf("Service %s created", service.GetName()))

	return updateStatusWithoutRepository(ctx, su, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonServiceCreated,
		Message:            fmt.Sprintf("Service %s created", service.GetName()),
	})
}

func updateService(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function, oldService corev1.Service, newService corev1.Service) error {
	svc := oldService.DeepCopy()

	// manually change fields that interest us, as clusterIP is immutable
	svc.Spec.Ports = newService.Spec.Ports
	svc.Spec.Selector = newService.Spec.Selector
	svc.Spec.Type = newService.Spec.Type

	svc.ObjectMeta.Labels = newService.GetLabels()

	log.Info(fmt.Sprintf("Updating Service %s", svc.GetName()))
	if err := su.client.Update(ctx, svc); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update Service with name %s", svc.GetName()))
		return err
	}
	log.Info(fmt.Sprintf("Service %s updated", svc.GetName()))

	return updateStatusWithoutRepository(ctx, su, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonServiceUpdated,
		Message:            fmt.Sprintf("Service %s updated", svc.GetName()),
	})
}

func deleteExcessServices(ctx context.Context, client resource.Client, instance *serverlessv1alpha1.Function, log logr.Logger, services []corev1.Service) error {
	// services do not support deletecollection
	// you can check this by `kubectl api-resources -o wide | grep services`
	// also https://github.com/kubernetes/kubernetes/issues/68468#issuecomment-419981870

	log.Info("Deleting excess Services")

	for i := range services {
		svc := services[i]
		if svc.GetName() == instance.GetName() {
			continue
		}

		log.Info(fmt.Sprintf("Deleting Service %s", svc.GetName()))

		err := client.Delete(ctx, &services[i])
		if err != nil {
			log.Error(err, fmt.Sprintf("Cannot delete excess Service %s", svc.GetName()))
			return err
		}
	}

	log.Info("Excess Services deleted")
	return nil
}

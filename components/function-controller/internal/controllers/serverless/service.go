package serverless

import (
	"context"
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func stateFnCheckService(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	err := r.client.ListByLabel(
		ctx,
		s.instance.GetNamespace(),
		internalFunctionLabels(s.instance),
		&s.services)

	if err != nil {
		return nil, errors.Wrap(err, "while listing services")
	}

	expectedSvc := s.buildService()

	if len(s.services.Items) == 0 {
		return buildStateFnCreateNewService(expectedSvc), nil
	}

	if len(s.services.Items) > 1 {
		return stateFnDeleteServices, nil
	}

	if s.svcChanged(expectedSvc) {
		return buildStateFnUpdateService(expectedSvc), nil
	}

	return stateFnCheckScaling, nil
}

func buildStateFnUpdateService(newService corev1.Service) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {

		svc := &s.services.Items[0]

		// manually change fields that interest us, as clusterIP is immutable
		svc.Spec.Ports = newService.Spec.Ports
		svc.Spec.Selector = newService.Spec.Selector
		svc.Spec.Type = newService.Spec.Type

		svc.ObjectMeta.Labels = newService.GetLabels()

		r.log.Info(fmt.Sprintf("Updating Service %s", svc.GetName()))

		err := r.client.Update(ctx, svc)
		if err != nil {
			condition := serverlessv1alpha2.Condition{
				Type:               serverlessv1alpha2.ConditionRunning,
				Status:             corev1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             serverlessv1alpha2.ConditionReasonServiceFailed,
				Message:            fmt.Sprintf("Service %s update error: %s", svc.GetName(), err.Error()),
			}
			return buildStatusUpdateStateFnWithCondition(condition), nil
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonServiceUpdated,
			Message:            fmt.Sprintf("Service %s updated", svc.GetName()),
		}

		return buildStatusUpdateStateFnWithCondition(condition), nil
	}
}

func buildStateFnCreateNewService(svc corev1.Service) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
		r.log.Info(fmt.Sprintf("Creating Service %s", svc.GetName()))

		err := r.client.CreateWithReference(ctx, &s.instance, &svc)
		if err != nil {
			condition := serverlessv1alpha2.Condition{
				Type:               serverlessv1alpha2.ConditionRunning,
				Status:             corev1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             serverlessv1alpha2.ConditionReasonServiceFailed,
				Message:            fmt.Sprintf("Service %s create error: %s", svc.GetName(), err.Error()),
			}
			return buildStatusUpdateStateFnWithCondition(condition), nil
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonServiceCreated,
			Message:            fmt.Sprintf("Service %s created", svc.GetName()),
		}

		return buildStatusUpdateStateFnWithCondition(condition), nil
	}
}

func stateFnDeleteServices(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	// services do not support deletecollection
	// you can check this by `kubectl api-resources -o wide | grep services`
	// also https://github.com/kubernetes/kubernetes/issues/68468#issuecomment-419981870

	r.log.Info("deleting Services")

	for i := range s.services.Items {
		svc := s.services.Items[i]
		if svc.GetName() == s.instance.GetName() {
			continue
		}

		r.log.Info(fmt.Sprintf("deleting Service %s", svc.GetName()))

		// TODO consider implementing mechanism to collect errors
		err := r.client.Delete(ctx, &s.services.Items[i])
		if err != nil {
			return nil, errors.Wrap(err, "while deleting service")
		}
	}

	return nil, nil
}

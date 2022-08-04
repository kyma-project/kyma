package serverless

import (
	"context"
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func stateFnCheckService(ctx context.Context, r *reconciler, s *systemState) stateFn {
	r.err = r.client.ListByLabel(
		ctx,
		s.instance.GetNamespace(),
		s.internalFunctionLabels(),
		&s.services)

	if r.err != nil {
		return nil
	}

	expectedSvc := s.buildService()

	if len(s.services.Items) == 0 {
		return buildStateFnCreateNewService(expectedSvc)
	}

	if len(s.services.Items) > 1 {
		return stateFnDeleteServices
	}

	if s.svcChanged(expectedSvc) {
		return buildStateFnUpdateService(expectedSvc)
	}

	return stateFnCheckHPA
}

func buildStateFnUpdateService(newService corev1.Service) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) stateFn {

		svc := &s.services.Items[0]

		// manually change fields that interest us, as clusterIP is immutable
		svc.Spec.Ports = newService.Spec.Ports
		svc.Spec.Selector = newService.Spec.Selector
		svc.Spec.Type = newService.Spec.Type

		svc.ObjectMeta.Labels = newService.GetLabels()

		r.log.Info(fmt.Sprintf("Updating Service %s", svc.GetName()))

		r.err = r.client.Update(ctx, svc)
		if r.err != nil {
			return nil
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonServiceUpdated,
			Message:            fmt.Sprintf("Service %s updated", svc.GetName()),
		}

		return buildStateFnUpdateStateFnFunctionCondition(condition)
	}
}

func buildStateFnCreateNewService(svc corev1.Service) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		r.log.Info(fmt.Sprintf("Creating Service %s", svc.GetName()))

		r.err = r.client.CreateWithReference(ctx, &s.instance, &svc)
		if r.err != nil {
			return nil
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonServiceCreated,
			Message:            fmt.Sprintf("Service %s created", svc.GetName()),
		}

		return buildStateFnUpdateStateFnFunctionCondition(condition)
	}
}

func stateFnDeleteServices(ctx context.Context, r *reconciler, s *systemState) stateFn {
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
		r.err = r.client.Delete(ctx, &s.services.Items[i])
		if r.err != nil {
			return nil
		}
	}

	return nil
}

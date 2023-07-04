package serverless

import (
	"context"
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
)

func stateFnCheckScaling(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	namespace := s.instance.GetNamespace()
	labels := internalFunctionLabels(s.instance)

	err := r.client.ListByLabel(ctx, namespace, labels, &s.hpas)
	if err != nil {
		return nil, errors.Wrap(err, "while listing HPAs")
	}

	if !isScalingEnabled(&s.instance) {
		return stateFnCheckReplicas(ctx, r, s)
	}

	return stateFnCheckHPA(ctx, r, s)
}

func stateFnCheckReplicas(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	numHpa := len(s.hpas.Items)

	if numHpa > 0 {
		return stateFnDeleteAllHorizontalPodAutoscalers, nil
	}

	return stateFnUpdateDeploymentStatus, nil
}

func stateFnCheckHPA(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	if !s.hpaEqual(r.cfg.fn.TargetCPUUtilizationPercentage) {
		return stateFnUpdateDeploymentStatus, nil
	}

	numHpa := len(s.hpas.Items)
	expectedHPA := s.buildHorizontalPodAutoscaler(r.cfg.fn.TargetCPUUtilizationPercentage)

	if numHpa == 0 {
		if !equalInt32Pointer(s.instance.Spec.ScaleConfig.MinReplicas, s.instance.Spec.ScaleConfig.MaxReplicas) {
			return buildStateFnCreateHorizontalPodAutoscaler(expectedHPA), nil
		}
		return nil, nil
	}

	if numHpa > 1 {
		return stateFnDeleteAllHorizontalPodAutoscalers, nil
	}

	if numHpa == 1 && !isScalingEnabled(&s.instance) {
		// this case is when we previously created HPA with maxReplicas > minReplicas, but now user changed
		// function spec and NOW maxReplicas == minReplicas, so hpa is not needed anymore
		return stateFnDeleteAllHorizontalPodAutoscalers, nil
	}

	if !s.equalHorizontalPodAutoscalers(expectedHPA) {
		return buildStateFnUpdateHorizontalPodAutoscaler(expectedHPA), nil
	}

	return nil, nil
}

func buildStateFnCreateHorizontalPodAutoscaler(hpa autoscalingv1.HorizontalPodAutoscaler) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
		r.log.Info("Creating HorizontalPodAutoscaler")

		err := r.client.CreateWithReference(ctx, &s.instance, &hpa)
		if err != nil {
			return nil, errors.Wrap(err, "while creating HPA")
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonHorizontalPodAutoscalerCreated,
			Message:            fmt.Sprintf("HorizontalPodAutoscaler %s created", hpa.GetName()),
		}

		return buildStatusUpdateStateFnWithCondition(condition), nil
	}
}

func stateFnDeleteAllHorizontalPodAutoscalers(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	r.log.Info("Deleting all HorizontalPodAutoscalers attached to function")
	selector := apilabels.SelectorFromSet(internalFunctionLabels(s.instance))

	err := r.client.DeleteAllBySelector(ctx, &autoscalingv1.HorizontalPodAutoscaler{}, s.instance.GetNamespace(), selector)
	return nil, errors.Wrap(err, "while deleting HPAs")
}

func buildStateFnUpdateHorizontalPodAutoscaler(expectd autoscalingv1.HorizontalPodAutoscaler) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
		hpa := &s.hpas.Items[0]

		hpa.Spec = expectd.Spec
		hpa.Labels = expectd.GetLabels()

		hpaName := hpa.GetName()

		r.log.Info(fmt.Sprintf("Updating HorizontalPodAutoscaler %s", hpaName))

		err := r.client.Update(ctx, hpa)
		if err != nil {
			return nil, errors.Wrap(err, "while updating HPA")
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonHorizontalPodAutoscalerUpdated,
			Message:            fmt.Sprintf("HorizontalPodAutoscaler %s updated", hpaName),
		}
		return buildStatusUpdateStateFnWithCondition(condition), nil
	}
}

package serverless

import (
	"context"
	"fmt"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
)

func stateFnCheckHPA(ctx context.Context, r *reconciler, s *systemState) stateFn {
	namespace := s.instance.GetNamespace()
	labels := s.instance.GenerateInternalLabels()

	r.err = r.client.ListByLabel(ctx, namespace, labels, &s.hpas)
	if r.err != nil {
		return nil
	}

	if !s.hpaEqual(r.cfg.fn.TargetCPUUtilizationPercentage) {
		return stateFnUpdateDeploymentStatus
	}

	numHpa := len(s.hpas.Items)
	expectedHPA := buildFunctionHPA(s.instance, s.deployments.Items[0].GetName(), r.cfg.fn.TargetCPUUtilizationPercentage)

	if numHpa == 0 {
		if !equalInt32Pointer(s.instance.Spec.MinReplicas, s.instance.Spec.MaxReplicas) {
			return buildStateFnCreateHorizontalPodAutoscaler(expectedHPA)
		}
		return nil
	}

	if numHpa > 1 {
		return stateFnDeleteAllHorizontalPodAutoscalers
	}

	if numHpa == 1 && equalInt32Pointer(s.instance.Spec.MinReplicas, s.instance.Spec.MaxReplicas) {
		// this case is when we previously created HPA with maxReplicas > minReplicas, but now user changed
		// function spec and NOW maxReplicas == minReplicas, so hpa is not needed anymore
		return stateFnDeleteAllHorizontalPodAutoscalers
	}

	if !equalHorizontalPodAutoscalers(s.hpas.Items[0], expectedHPA) {
		return buildStateFnUpdateHorizontalPodAutoscaler(expectedHPA)
	}

	return nil
}

func buildStateFnCreateHorizontalPodAutoscaler(hpa autoscalingv1.HorizontalPodAutoscaler) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		r.log.Info("Creating HorizontalPodAutoscaler")

		r.err = r.client.CreateWithReference(ctx, &s.instance, &hpa)
		if r.err != nil {
			return nil
		}

		condition := serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonHorizontalPodAutoscalerCreated,
			Message:            fmt.Sprintf("HorizontalPodAutoscaler %s created", hpa.GetName()),
		}

		return buildStateFnUpdateStateFnFunctionCondition(condition)
	}
}

func stateFnDeleteAllHorizontalPodAutoscalers(ctx context.Context, r *reconciler, s *systemState) stateFn {
	r.log.Info("Deleting all HorizontalPodAutoscalers attached to function")
	selector := apilabels.SelectorFromSet(s.instance.GenerateInternalLabels())

	r.err = r.client.DeleteAllBySelector(ctx, &autoscalingv1.HorizontalPodAutoscaler{}, s.instance.GetNamespace(), selector)
	if r.err != nil {
		return nil
	}

	return nil
}

func buildStateFnUpdateHorizontalPodAutoscaler(expectd autoscalingv1.HorizontalPodAutoscaler) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		hpa := &s.hpas.Items[0]

		hpa.Spec = expectd.Spec
		hpa.Labels = expectd.GetLabels()

		hpaName := hpa.GetName()

		r.log.Info(fmt.Sprintf("Updating HorizontalPodAutoscaler %s", hpaName))

		r.err = r.client.Update(ctx, hpa)
		if r.err != nil {
			return nil
		}

		condition := serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonHorizontalPodAutoscalerUpdated,
			Message:            fmt.Sprintf("HorizontalPodAutoscaler %s updated", hpaName),
		}

		return buildStateFnUpdateStateFnFunctionCondition(condition)
	}
}

func buildFunctionHPA(instance serverlessv1alpha1.Function, deployName string, targetCPUUtilizationPercentage int32) autoscalingv1.HorizontalPodAutoscaler {
	minReplicas, maxReplicas := instance.GetDefaultReplicas()
	deploymentName := deployName
	return autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
			Labels:       instance.GetMergedLables(),
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       deploymentName,
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
			MinReplicas:                    &minReplicas,
			MaxReplicas:                    maxReplicas,
			TargetCPUUtilizationPercentage: &targetCPUUtilizationPercentage,
		},
	}
}

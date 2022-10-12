package serverless

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
)

func stateFnInlineCheckSources(ctx context.Context, r *reconciler, s *systemState) stateFn {

	labels := s.internalFunctionLabels()

	r.err = r.client.ListByLabel(ctx, s.instance.GetNamespace(), labels, &s.configMaps)
	if r.err != nil {
		return nil
	}

	r.err = r.client.ListByLabel(ctx, s.instance.GetNamespace(), labels, &s.deployments)
	if r.err != nil {
		return nil
	}

	srcChanged := s.inlineFnSrcChanged(r.cfg.docker.PullAddress)
	if !srcChanged {
		expectedJob := s.buildInlineJob(s.configMaps.Items[0].GetName(), r.cfg)
		return buildStateFnCheckImageJob(expectedJob)
	}

	cfgMapCount := len(s.configMaps.Items)

	// TODO create issue to refactor the way how function controller is handling status
	next := stateFnInlineDeleteConfigMap

	if cfgMapCount == 0 {
		next = stateFnInlineCreateConfigMap
	}

	if cfgMapCount == 1 {
		next = stateFnInlineUpdateConfigMap
	}

	return next
}

func stateFnInlineDeleteConfigMap(ctx context.Context, r *reconciler, s *systemState) stateFn {
	r.log.Info("delete all ConfigMaps")

	labels := s.internalFunctionLabels()
	selector := apilabels.SelectorFromSet(labels)

	r.err = r.client.DeleteAllBySelector(ctx, &corev1.ConfigMap{}, s.instance.GetNamespace(), selector)
	if r.err != nil {
		return nil
	}

	return nil
}

func stateFnInlineCreateConfigMap(ctx context.Context, r *reconciler, s *systemState) stateFn {
	configMap := s.buildConfigMap()

	r.err = r.client.CreateWithReference(ctx, &s.instance, &configMap)
	if r.err != nil {
		return nil
	}

	currentCondition := serverlessv1alpha2.Condition{
		Type:               serverlessv1alpha2.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha2.ConditionReasonConfigMapCreated,
		Message:            fmt.Sprintf("ConfigMap %s created", configMap.GetName()),
	}

	return buildStatusUpdateStateFnWithCondition(currentCondition)
}

func stateFnInlineUpdateConfigMap(ctx context.Context, r *reconciler, s *systemState) stateFn {
	expectedConfigMap := s.buildConfigMap()

	s.configMaps.Items[0].Data = expectedConfigMap.Data
	s.configMaps.Items[0].Labels = expectedConfigMap.Labels

	cmName := s.configMaps.Items[0].GetName()

	r.log.Info(fmt.Sprintf("updating ConfigMap %s", cmName))

	r.err = r.client.Update(ctx, &s.configMaps.Items[0])
	if r.err != nil {
		return nil
	}

	condition := serverlessv1alpha2.Condition{
		Type:               serverlessv1alpha2.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha2.ConditionReasonConfigMapUpdated,
		Message:            fmt.Sprintf("Updated ConfigMap: %q", cmName),
	}

	return buildStatusUpdateStateFnWithCondition(condition)
}

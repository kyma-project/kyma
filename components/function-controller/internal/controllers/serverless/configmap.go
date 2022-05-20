package serverless

import (
	"context"
	"fmt"

	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
)

func stateFnInlineCheckSources(ctx context.Context, r *reconciler, s *systemState) stateFn {

	labels := s.instance.GenerateInternalLabels()

	r.err = r.client.ListByLabel(ctx, s.instance.GetNamespace(), labels, &s.configMaps)
	if r.err != nil {
		return nil
	}

	r.err = r.client.ListByLabel(ctx, s.instance.GetNamespace(), labels, &s.deployments)
	if r.err != nil {
		return nil
	}

	srcChanged := s.inlineSourceChanged(r.cfg.docker.PullAddress)
	if !srcChanged {
		expectedJob := buildJob(s.instance, s.configMaps.Items[0].GetName(), r.cfg)
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

	labels := s.instance.GenerateInternalLabels()
	selector := apilabels.SelectorFromSet(labels)

	r.err = r.client.DeleteAllBySelector(ctx, &corev1.ConfigMap{}, s.instance.GetNamespace(), selector)
	if r.err != nil {
		return nil
	}

	return nil
}

func stateFnInlineCreateConfigMap(ctx context.Context, r *reconciler, s *systemState) stateFn {
	configMap := buildInlineSourceConfigMap(s.instance)

	r.err = r.client.CreateWithReference(ctx, &s.instance, &configMap)
	if r.err != nil {
		return nil
	}

	currentCondition := serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapCreated,
		Message:            fmt.Sprintf("ConfigMap %s created", configMap.GetName()),
	}

	return buildStateFnUpdateStateFnFunctionCondition(currentCondition)
}

func stateFnInlineUpdateConfigMap(ctx context.Context, r *reconciler, s *systemState) stateFn {
	expectedConfigMap := buildInlineSourceConfigMap(s.instance)

	s.configMaps.Items[0].Data = expectedConfigMap.Data
	s.configMaps.Items[0].Labels = expectedConfigMap.Labels

	cmName := s.configMaps.Items[0].GetName()

	r.log.Info(fmt.Sprintf("updating ConfigMap %s", cmName))

	r.err = r.client.Update(ctx, &s.configMaps.Items[0])
	if r.err != nil {
		return nil
	}

	condition := serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapUpdated,
		Message:            fmt.Sprintf("Updated ConfigMap: %q", cmName),
	}

	return buildStateFnUpdateStateFnFunctionCondition(condition)
}

func buildInlineSourceConfigMap(instance serverlessv1alpha1.Function) corev1.ConfigMap {
	rtm := fnRuntime.GetRuntime(instance.Spec.Runtime)
	data := map[string]string{
		FunctionSourceKey: instance.Spec.Source,
		FunctionDepsKey:   rtm.SanitizeDependencies(instance.Spec.Deps),
	}
	labels := instance.GetMergedLables()

	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       labels,
			GenerateName: fmt.Sprintf("%s-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
		},
		Data: data,
	}
}

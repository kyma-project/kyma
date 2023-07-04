package serverless

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
)

func stateFnInlineCheckSources(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {

	labels := internalFunctionLabels(s.instance)

	err := r.client.ListByLabel(ctx, s.instance.GetNamespace(), labels, &s.configMaps)
	if err != nil {
		return nil, errors.Wrap(err, "while listing configMaps")
	}

	err = r.client.ListByLabel(ctx, s.instance.GetNamespace(), labels, &s.deployments)
	if err != nil {
		return nil, errors.Wrap(err, "while listing deployments")
	}

	srcChanged := s.inlineFnSrcChanged(r.cfg.docker.PullAddress)
	if !srcChanged {
		expectedJob := s.buildJob(s.configMaps.Items[0].GetName(), r.cfg)
		return buildStateFnCheckImageJob(expectedJob), nil
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

	return next, nil
}

func stateFnInlineDeleteConfigMap(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	r.log.Info("delete all ConfigMaps")

	labels := internalFunctionLabels(s.instance)
	selector := apilabels.SelectorFromSet(labels)

	err := r.client.DeleteAllBySelector(ctx, &corev1.ConfigMap{}, s.instance.GetNamespace(), selector)
	if err != nil {
		return nil, errors.Wrap(err, "while deleting configMaps")
	}

	return nil, nil
}

func stateFnInlineCreateConfigMap(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	configMap := s.buildConfigMap()

	err := r.client.CreateWithReference(ctx, &s.instance, &configMap)
	if err != nil {
		return nil, errors.Wrap(err, "while creating configMaps")
	}

	currentCondition := serverlessv1alpha2.Condition{
		Type:               serverlessv1alpha2.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha2.ConditionReasonConfigMapCreated,
		Message:            fmt.Sprintf("ConfigMap %s created", configMap.GetName()),
	}

	return buildStatusUpdateStateFnWithCondition(currentCondition), nil
}

func stateFnInlineUpdateConfigMap(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	expectedConfigMap := s.buildConfigMap()

	s.configMaps.Items[0].Data = expectedConfigMap.Data
	s.configMaps.Items[0].Labels = expectedConfigMap.Labels

	cmName := s.configMaps.Items[0].GetName()

	r.log.Info(fmt.Sprintf("updating ConfigMap %s", cmName))

	err := r.client.Update(ctx, &s.configMaps.Items[0])
	if err != nil {
		return nil, errors.Wrap(err, "while updating configMap")
	}

	condition := serverlessv1alpha2.Condition{
		Type:               serverlessv1alpha2.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha2.ConditionReasonConfigMapUpdated,
		Message:            fmt.Sprintf("Updated ConfigMap: %q", cmName),
	}

	return buildStatusUpdateStateFnWithCondition(condition), nil
}

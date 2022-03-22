package serverless

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func (r *FunctionReconciler) createConfigMap(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, rtm runtime.Runtime) error {
	configMap := r.buildConfigMap(instance, rtm)

	log.Info("CreateWithReference ConfigMap")
	if err := r.client.CreateWithReference(ctx, instance, &configMap); err != nil {
		log.Error(err, "Cannot create ConfigMap")
		return err
	}
	log.Info(fmt.Sprintf("ConfigMap %s created", configMap.GetName()))

	return r.updateStatusWithoutRepository(ctx, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapCreated,
		Message:            fmt.Sprintf("ConfigMap %s created", configMap.GetName()),
	})
}

func (r *FunctionReconciler) updateConfigMap(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, rtm runtime.Runtime, configMap corev1.ConfigMap) error {
	newConfigMap := configMap.DeepCopy()
	expectedConfigMap := r.buildConfigMap(instance, rtm)

	newConfigMap.Data = expectedConfigMap.Data
	newConfigMap.Labels = expectedConfigMap.Labels

	log.Info(fmt.Sprintf("Updating ConfigMap %s", configMap.GetName()))
	if err := r.client.Update(ctx, newConfigMap); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update ConfigMap %s", newConfigMap.GetName()))
		return err
	}
	log.Info(fmt.Sprintf("ConfigMap %s updated", configMap.GetName()))

	return r.updateStatusWithoutRepository(ctx, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapUpdated,
		Message:            fmt.Sprintf("ConfigMap %s updated", configMap.GetName()),
	})
}

func (r *FunctionReconciler) deleteAllConfigMaps(ctx context.Context, instance *serverlessv1alpha1.Function, log logr.Logger) error {
	log.Info("Deleting all ConfigMaps")
	selector := apilabels.SelectorFromSet(internalFunctionLabels(instance))
	if err := r.client.DeleteAllBySelector(ctx, &corev1.ConfigMap{}, instance.GetNamespace(), selector); err != nil {
		log.Error(err, "Cannot delete all ConfigMaps")
		return err
	}

	log.Info("All underlying ConfigMaps deleted")
	return nil
}

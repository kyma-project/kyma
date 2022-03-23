package serverless

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func createConfigMap(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function, rtm runtime.Runtime) error {
	configMap := buildConfigMap(instance, rtm)

	log.Info("CreateWithReference ConfigMap")
	if err := su.client.CreateWithReference(ctx, instance, &configMap); err != nil {
		log.Error(err, "Cannot create ConfigMap")
		return err
	}
	log.Info(fmt.Sprintf("ConfigMap %s created", configMap.GetName()))

	return updateStatusWithoutRepository(ctx, su, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapCreated,
		Message:            fmt.Sprintf("ConfigMap %s created", configMap.GetName()),
	})
}

func updateConfigMap(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function, rtm runtime.Runtime, configMap corev1.ConfigMap) error {
	newConfigMap := configMap.DeepCopy()
	expectedConfigMap := buildConfigMap(instance, rtm)

	newConfigMap.Data = expectedConfigMap.Data
	newConfigMap.Labels = expectedConfigMap.Labels

	log.Info(fmt.Sprintf("Updating ConfigMap %s", configMap.GetName()))
	if err := su.client.Update(ctx, newConfigMap); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update ConfigMap %s", newConfigMap.GetName()))
		return err
	}
	log.Info(fmt.Sprintf("ConfigMap %s updated", configMap.GetName()))

	return updateStatusWithoutRepository(ctx, su, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapUpdated,
		Message:            fmt.Sprintf("ConfigMap %s updated", configMap.GetName()),
	})
}

func isOnConfigMapChange(instance *serverlessv1alpha1.Function, rtm runtime.Runtime, configMaps []corev1.ConfigMap, deployments []appsv1.Deployment, dockerConfig DockerConfig) bool {
	image := buildImageAddress(instance, dockerConfig.PullAddress)
	configurationStatus := getConditionStatus(instance.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)

	if len(deployments) == 1 &&
		len(configMaps) == 1 &&
		deployments[0].Spec.Template.Spec.Containers[0].Image == image &&
		configurationStatus != corev1.ConditionUnknown &&
		mapsEqual(configMaps[0].Labels, functionLabels(instance)) {
		return false
	}

	return !(len(configMaps) == 1 &&
		instance.Spec.Source == configMaps[0].Data[FunctionSourceKey] &&
		rtm.SanitizeDependencies(instance.Spec.Deps) == configMaps[0].Data[FunctionDepsKey] &&
		configurationStatus == corev1.ConditionTrue &&
		mapsEqual(configMaps[0].Labels, functionLabels(instance)))
}

func (r *FunctionReconciler) onConfigMapChange(ctx context.Context, su *statusUpdater, log logr.Logger, instance *serverlessv1alpha1.Function, rtm runtime.Runtime, configMaps []corev1.ConfigMap) error {
	configMapsLen := len(configMaps)

	switch configMapsLen {
	case 0:
		return createConfigMap(ctx, su, log, instance, rtm)
	case 1:
		return updateConfigMap(ctx, su, log, instance, rtm, configMaps[0])
	default:
		return deleteAllConfigMaps(ctx, su.client, instance, log)
	}
}

func deleteAllConfigMaps(ctx context.Context, client resource.Client, instance *serverlessv1alpha1.Function, log logr.Logger) error {
	log.Info("Deleting all ConfigMaps")
	selector := apilabels.SelectorFromSet(internalFunctionLabels(instance))
	if err := client.DeleteAllBySelector(ctx, &corev1.ConfigMap{}, instance.GetNamespace(), selector); err != nil {
		log.Error(err, "Cannot delete all ConfigMaps")
		return err
	}

	log.Info("All underlying ConfigMaps deleted")
	return nil
}

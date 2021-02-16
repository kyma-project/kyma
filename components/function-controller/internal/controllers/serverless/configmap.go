package serverless

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func (r *FunctionReconciler) createConfigMap(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, rtm runtime.Runtime) (ctrl.Result, error) {
	configMap := r.buildConfigMap(instance, rtm)

	log.Info("CreateWithReference ConfigMap")
	if err := r.client.CreateWithReference(ctx, instance, &configMap); err != nil {
		log.Error(err, "Cannot create ConfigMap")
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("ConfigMap %s created", configMap.GetName()))

	return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapCreated,
		Message:            fmt.Sprintf("ConfigMap %s created", configMap.GetName()),
	})
}

func (r *FunctionReconciler) updateConfigMap(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, rtm runtime.Runtime, configMap corev1.ConfigMap) (ctrl.Result, error) {
	newConfigMap := configMap.DeepCopy()
	expectedConfigMap := r.buildConfigMap(instance, rtm)

	newConfigMap.Data = expectedConfigMap.Data
	newConfigMap.Labels = expectedConfigMap.Labels

	log.Info(fmt.Sprintf("Updating ConfigMap %s", configMap.GetName()))
	if err := r.client.Update(ctx, newConfigMap); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update ConfigMap %s", newConfigMap.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("ConfigMap %s updated", configMap.GetName()))

	return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapUpdated,
		Message:            fmt.Sprintf("ConfigMap %s updated", configMap.GetName()),
	})
}

func (r *FunctionReconciler) isOnConfigMapChange(instance *serverlessv1alpha1.Function, rtm runtime.Runtime, configMaps []corev1.ConfigMap, deployments []appsv1.Deployment, dockerConfig DockerConfig) bool {
	image := r.buildImageAddress(instance, dockerConfig.PullAddress)
	configurationStatus := r.getConditionStatus(instance.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)

	if len(deployments) == 1 &&
		len(configMaps) == 1 &&
		deployments[0].Spec.Template.Spec.Containers[0].Image == image &&
		configurationStatus != corev1.ConditionUnknown &&
		r.mapsEqual(configMaps[0].Labels, r.functionLabels(instance)) {
		return false
	}

	return !(len(configMaps) == 1 &&
		instance.Spec.Source == configMaps[0].Data[FunctionSourceKey] &&
		rtm.SanitizeDependencies(instance.Spec.Deps) == configMaps[0].Data[FunctionDepsKey] &&
		configurationStatus == corev1.ConditionTrue &&
		r.mapsEqual(configMaps[0].Labels, r.functionLabels(instance)))
}

func (r *FunctionReconciler) onConfigMapChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, rtm runtime.Runtime, configMaps []corev1.ConfigMap) (ctrl.Result, error) {
	configMapsLen := len(configMaps)

	switch configMapsLen {
	case 0:
		return r.createConfigMap(ctx, log, instance, rtm)
	case 1:
		return r.updateConfigMap(ctx, log, instance, rtm, configMaps[0])
	default:
		return r.deleteAllConfigMaps(ctx, instance, log)
	}
}

func (r *FunctionReconciler) deleteAllConfigMaps(ctx context.Context, instance *serverlessv1alpha1.Function, log logr.Logger) (ctrl.Result, error) {
	log.Info("Deleting all ConfigMaps")
	selector := apilabels.SelectorFromSet(r.internalFunctionLabels(instance))
	if err := r.client.DeleteAllBySelector(ctx, &corev1.ConfigMap{}, instance.GetNamespace(), selector); err != nil {
		log.Error(err, "Cannot delete all ConfigMaps")
		return ctrl.Result{}, err
	}

	log.Info("All underlying ConfigMaps deleted")
	return ctrl.Result{}, nil
}

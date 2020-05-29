package serverless

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func (r *FunctionReconciler) createConfigMap(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function) (ctrl.Result, error) {
	configMap := r.buildConfigMap(instance)

	log.Info("CreateWithReference ConfigMap")
	if err := r.client.CreateWithReference(ctx, instance, &configMap); err != nil {
		log.Error(err, "Cannot create ConfigMap")
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("ConfigMap %s created", configMap.GetName()))

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapCreated,
		Message:            fmt.Sprintf("ConfigMap %s created", configMap.GetName()),
	})
}

func (r *FunctionReconciler) updateConfigMap(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, configMap corev1.ConfigMap) (ctrl.Result, error) {
	newConfigMap := configMap.DeepCopy()
	expectedConfigMap := r.buildConfigMap(instance)

	newConfigMap.Data = expectedConfigMap.Data
	newConfigMap.Labels = expectedConfigMap.Labels

	log.Info(fmt.Sprintf("Updating ConfigMap %s", configMap.GetName()))
	if err := r.client.Update(ctx, newConfigMap); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update ConfigMap %s", newConfigMap.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("ConfigMap %s updated", configMap.GetName()))

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapUpdated,
		Message:            fmt.Sprintf("ConfigMap %s updated", configMap.GetName()),
	})
}

func (r *FunctionReconciler) isOnConfigMapChange(instance *serverlessv1alpha1.Function, configMaps []corev1.ConfigMap, deployments []appsv1.Deployment) bool {
	image := r.buildExternalImageAddress(instance)
	configurationStatus := r.getConditionStatus(instance.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)

	if len(deployments) == 1 && deployments[0].Spec.Template.Spec.Containers[0].Image == image && configurationStatus != corev1.ConditionUnknown {
		return false
	}

	return len(configMaps) != 1 ||
		instance.Spec.Source != configMaps[0].Data[configMapFunction] ||
		r.sanitizeDependencies(instance.Spec.Deps) != configMaps[0].Data[configMapDeps] ||
		configMaps[0].Data[configMapHandler] != configMapHandler ||
		configurationStatus != corev1.ConditionTrue
}

func (r *FunctionReconciler) onConfigMapChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, configMaps []corev1.ConfigMap) (ctrl.Result, error) {
	configMapsLen := len(configMaps)

	switch configMapsLen {
	case 0:
		return r.createConfigMap(ctx, log, instance)
	case 1:
		return r.updateConfigMap(ctx, log, instance, configMaps[0])
	default:
		log.Info("Too many ConfigMaps for function")
		return r.updateStatus(ctx, ctrl.Result{RequeueAfter: r.config.RequeueDuration}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonConfigMapError,
			Message:            "Too many ConfigMaps for function",
		})
	}
}

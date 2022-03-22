package serverless

import (
	"context"

	"github.com/go-logr/logr"
	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *FunctionReconciler) reconcileInlineFunction(ctx context.Context, instance *serverlessv1alpha1.Function, log logr.Logger) (ctrl.Result, error) {

	resources, err := r.fetchFunctionResources(ctx, instance, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	dockerConfig, err := r.readDockerConfig(ctx, instance)
	if err != nil {
		log.Error(err, "Cannot read Docker registry configuration")
		return ctrl.Result{}, err
	}
	rtmCfg := fnRuntime.GetRuntimeConfig(instance.Spec.Runtime)
	rtm := fnRuntime.GetRuntime(instance.Spec.Runtime)
	var result ctrl.Result

	switch {

	case r.isOnConfigMapChange(instance, rtm, resources.configMaps.Items, resources.deployments.Items, dockerConfig):
		return result, r.onConfigMapChange(ctx, log, instance, rtm, resources.configMaps.Items)

	case r.isOnJobChange(instance, rtmCfg, resources.jobs.Items, resources.deployments.Items, git.Options{}, dockerConfig):
		return r.onJobChange(ctx, log, instance, rtmCfg, resources.configMaps.Items[0].GetName(), resources.jobs.Items, dockerConfig)

	case r.isOnDeploymentChange(instance, rtmCfg, resources.deployments.Items, dockerConfig):
		return r.onDeploymentChange(ctx, log, instance, rtmCfg, resources.deployments.Items, dockerConfig)

	case r.isOnServiceChange(instance, resources.services.Items):
		return result, r.onServiceChange(ctx, log, instance, resources.services.Items)

	case r.isOnHorizontalPodAutoscalerChange(instance, resources.hpas.Items, resources.deployments.Items):
		return result, r.onHorizontalPodAutoscalerChange(ctx, log, instance, resources.hpas.Items, resources.deployments.Items[0].GetName())

	default:
		return r.updateDeploymentStatus(ctx, log, instance, resources.deployments.Items, corev1.ConditionTrue)
	}
}

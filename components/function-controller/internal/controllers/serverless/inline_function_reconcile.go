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

func (r *FunctionReconciler) reconcileInlineFunctionReconcile(ctx context.Context, instance *serverlessv1alpha1.Function, resources *functionResources, su *statusUpdater, log logr.Logger) (ctrl.Result, error) {

	dockerConfig, err := readDockerConfig(ctx, r.client, r.config, instance)
	if err != nil {
		log.Error(err, "Cannot read Docker registry configuration")
		return ctrl.Result{}, err
	}
	rtmCfg := fnRuntime.GetRuntimeConfig(instance.Spec.Runtime)
	rtm := fnRuntime.GetRuntime(instance.Spec.Runtime)
	var result ctrl.Result

	switch {
	case isOnConfigMapChange(instance, rtm, resources.configMaps.Items, resources.deployments.Items, dockerConfig):
		return result, r.onConfigMapChange(ctx, su, log, instance, rtm, resources.configMaps.Items)

	case isOnJobChange(instance, rtmCfg, resources.jobs.Items, resources.deployments.Items, git.Options{}, dockerConfig, r.config):
		return onJobChange(ctx, su, log, instance, rtmCfg, resources.configMaps.Items[0].GetName(), resources.jobs.Items, dockerConfig)

	case isOnDeploymentChange(instance, rtmCfg, resources.deployments.Items, dockerConfig, r.config):
		return onDeploymentChange(ctx, su, log, instance, rtmCfg, resources.deployments.Items, dockerConfig, r.config)

	case isOnServiceChange(instance, resources.services.Items):
		return result, onServiceChange(ctx, su, log, instance, resources.services.Items)

	case isOnHorizontalPodAutoscalerChange(instance, resources.hpas.Items, resources.deployments.Items, r.config):
		return result, onHorizontalPodAutoscalerChange(ctx, su, log, instance, resources.hpas.Items, resources.deployments.Items[0].GetName(), r.config)

	default:
		return updateDeploymentStatus(ctx, su, log, instance, resources.deployments.Items, corev1.ConditionTrue)
	}
}

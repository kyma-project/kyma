package serverless

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *FunctionReconciler) reconcileGitFunction(ctx context.Context, instance *serverlessv1alpha1.Function, log logr.Logger) (ctrl.Result, error) {

	resources, err := r.fetchFunctionResources(ctx, instance, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	dockerConfig, err := r.readDockerConfig(ctx, instance)
	if err != nil {
		log.Error(err, "Cannot read Docker registry configuration")
		return ctrl.Result{}, err
	}

	gitOptions, err := r.readGITOptions(ctx, instance)
	if err != nil {
		if updateErr := r.updateStatusWithoutRepository(ctx, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonSourceUpdateFailed,
			Message:            fmt.Sprintf("Reading git options failed: %v", err),
		}); updateErr != nil {
			log.Error(err, "Reading git options failed")
			return ctrl.Result{}, errors.Wrap(updateErr, "while updating status")
		}
		return ctrl.Result{}, err
	}

	revision, err := r.syncRevision(instance, gitOptions)
	if err != nil {
		result, errMsg := NextRequeue(err)
		// TODO: This return masks the error from r.syncRevision() and doesn't pass it to the controller. This should be fixed in a follow up PR.
		return result, r.updateStatusWithoutRepository(ctx, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonSourceUpdateFailed,
			Message:            errMsg,
		})
	}

	rtmCfg := fnRuntime.GetRuntimeConfig(instance.Spec.Runtime)
	var result ctrl.Result

	switch {
	case r.isOnSourceChange(instance, revision):
		return result, r.onSourceChange(ctx, instance, &serverlessv1alpha1.Repository{
			Reference: instance.Spec.Reference,
			BaseDir:   instance.Spec.Repository.BaseDir,
		}, revision)

	case r.isOnJobChange(instance, rtmCfg, resources.jobs.Items, resources.deployments.Items, gitOptions, dockerConfig):
		return r.onGitJobChange(ctx, log, instance, rtmCfg, resources.jobs.Items, gitOptions, dockerConfig)

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

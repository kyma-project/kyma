package serverless

import (
	"context"
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// Progressing:
	// NewRSAvailableReason is added in a deployment when its newest replica set is made available
	// ie. the number of new pods that have passed readiness checks and run for at least minReadySeconds
	// is at least the minimum available pods that need to run for the deployment.
	NewRSAvailableReason = "NewReplicaSetAvailable"

	// Available:
	// MinimumReplicasAvailable is added in a deployment when it has its minimum replicas required available.
	MinimumReplicasAvailable   = "MinimumReplicasAvailable"
	MinimumReplicasUnavailable = "MinimumReplicasUnavailable"
)

func stateFnCheckDeployments(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	labels := internalFunctionLabels(s.instance)

	err := r.client.ListByLabel(ctx, s.instance.GetNamespace(), labels, &s.deployments)
	if err != nil {
		return nil, errors.Wrap(err, "while Listing deployments")
	}

	if err = ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context error")
	}

	args := buildDeploymentArgs{
		DockerPullAddress:      r.cfg.docker.PullAddress,
		TraceCollectorEndpoint: r.cfg.fn.TraceCollectorEndpoint,
		PublisherProxyAddress:  r.cfg.fn.PublisherProxyAddress,
		ImagePullAccountName:   r.cfg.fn.ImagePullAccountName,
	}

	expectedDeployment := s.buildDeployment(args, r.cfg.fn.ResourceConfiguration.Function.Resources)
	if len(s.deployments.Items) == 0 {
		return buildStateFnCreateDeployment(expectedDeployment), nil
	}

	if len(s.deployments.Items) > 1 {
		return stateFnDeleteDeployments, nil
	}

	if !equalDeployments(s.deployments.Items[0], expectedDeployment) {
		return buildStateFnUpdateDeployment(expectedDeployment.Spec, expectedDeployment.Labels), nil
	}
	return stateFnCheckService, nil
}

func buildStateFnCreateDeployment(d appsv1.Deployment) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
		err := r.client.CreateWithReference(ctx, &s.instance, &d)
		if err != nil {
			return nil, errors.Wrap(err, "while creating deployment")
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonDeploymentCreated,
			Message:            fmt.Sprintf("Deployment %s created", d.GetName()),
		}

		return buildStatusUpdateStateFnWithCondition(condition), nil
	}
}

func stateFnDeleteDeployments(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	r.log.Info("deleting function")

	labels := internalFunctionLabels(s.instance)
	selector := apilabels.SelectorFromSet(labels)

	err := r.client.DeleteAllBySelector(ctx, &appsv1.Deployment{}, s.instance.GetNamespace(), selector)
	return nil, errors.Wrap(err, "while deleting delpoyments")
}

func buildStateFnUpdateDeployment(expectedSpec appsv1.DeploymentSpec, expectedLabels map[string]string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {

		s.deployments.Items[0].Spec = expectedSpec
		s.deployments.Items[0].Labels = expectedLabels
		deploymentName := s.deployments.Items[0].GetName()

		r.log.Info(fmt.Sprintf("updating Deployment %s", deploymentName))

		err := r.client.Update(ctx, &s.deployments.Items[0])
		if err != nil {
			return nil, errors.Wrap(err, "while updating deployment")
		}

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonDeploymentUpdated,
			Message:            fmt.Sprintf("Deployment %s updated", deploymentName),
		}

		return buildStatusUpdateStateFnWithCondition(condition), nil
	}
}

func stateFnUpdateDeploymentStatus(ctx context.Context, r *reconciler, s *systemState) (stateFn, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context error")
	}

	deploymentName := s.deployments.Items[0].GetName()

	// ready deployment
	if s.isDeploymentReady() {
		r.log.Info(fmt.Sprintf("deployment ready %q", deploymentName))

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionRunning,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonDeploymentReady,
			Message:            fmt.Sprintf("Deployment %s is ready", deploymentName),
		}

		r.result = ctrl.Result{
			RequeueAfter: r.cfg.fn.FunctionReadyRequeueDuration,
		}

		return buildStatusUpdateStateFnWithCondition(condition), nil
	}

	// unhealthy deployment
	if s.hasDeploymentConditionFalseStatusWithReason(appsv1.DeploymentAvailable, MinimumReplicasUnavailable) {
		r.log.Info(fmt.Sprintf("deployment unhealthy: %q", deploymentName))

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonMinReplicasNotAvailable,
			Message:            fmt.Sprintf("Minimum replcas not available for deployment %s", deploymentName),
		}

		return buildStatusUpdateStateFnWithCondition(condition), nil
	}

	// deployment not ready
	if s.hasDeploymentConditionTrueStatus(appsv1.DeploymentProgressing) {
		r.log.Info(fmt.Sprintf("deployment %q not ready", deploymentName))

		condition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonDeploymentWaiting,
			Message:            fmt.Sprintf("Deployment %s is not ready yet", deploymentName),
		}

		return buildStatusUpdateStateFnWithCondition(condition), nil
	}

	// deployment failed
	r.log.Info(fmt.Sprintf("deployment %q failed", deploymentName))

	yamlConditions, err := yaml.Marshal(s.deployments.Items[0].Status.Conditions)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing deployment status")
	}

	msg := fmt.Sprintf("Deployment %s failed with condition: \n%s", deploymentName, yamlConditions)

	condition := serverlessv1alpha2.Condition{
		Type:               serverlessv1alpha2.ConditionRunning,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha2.ConditionReasonDeploymentFailed,
		Message:            msg,
	}

	return buildStatusUpdateStateFnWithCondition(condition), nil
}

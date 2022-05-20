package serverless

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
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

func stateFnCheckDeployments(ctx context.Context, r *reconciler, s *systemState) stateFn {
	labels := s.instance.GenerateInternalLabels()

	r.err = r.client.ListByLabel(ctx, s.instance.GetNamespace(), labels, &s.deployments)
	if r.err != nil {
		return nil
	}

	if r.err = ctx.Err(); r.err != nil {
		return nil
	}

	args := buildDeploymentArgs{
		DockerPullAddress:     r.cfg.docker.PullAddress,
		JaegerServiceEndpoint: r.cfg.fn.JaegerServiceEndpoint,
		PublisherProxyAddress: r.cfg.fn.PublisherProxyAddress,
		ImagePullAccountName:  r.cfg.fn.ImagePullAccountName,
	}

	expectedDeployment := buildFunctionDeployment(s.instance, args)

	if len(s.deployments.Items) == 0 {
		return buildStateFnCreateDeployment(expectedDeployment)
	}

	if len(s.deployments.Items) > 1 {
		return stateFnDeleteDeployments
	}

	deploymentChanged := functionDeploymentChanged(s.deployments.Items[0], expectedDeployment, isScalingEnabled(&s.instance))

	if !deploymentChanged {
		return stateFnCheckService
	}
	if !equalDeployments(s.deployments.Items[0], expectedDeployment, isScalingEnabled(&s.instance)) {
		return buildStateFnUpdateDeployment(expectedDeployment.Spec, expectedDeployment.Labels)
	}

	return stateFnUpdateDeploymentStatus
}

func buildStateFnCreateDeployment(d appsv1.Deployment) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		r.err = r.client.CreateWithReference(ctx, &s.instance, &d)
		if r.err != nil {
			return nil
		}

		condition := serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonDeploymentCreated,
			Message:            fmt.Sprintf("Deployment %s created", d.GetName()),
		}

		return buildStateFnUpdateStateFnFunctionCondition(condition)
	}
}

func stateFnDeleteDeployments(ctx context.Context, r *reconciler, s *systemState) stateFn {
	r.log.Info("deleting function")

	labels := s.instance.GenerateInternalLabels()
	selector := apilabels.SelectorFromSet(labels)

	r.err = r.client.DeleteAllBySelector(ctx, &appsv1.Deployment{}, s.instance.GetNamespace(), selector)
	return nil
}

func buildStateFnUpdateDeployment(expectedSpec appsv1.DeploymentSpec, expectedLabels map[string]string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) stateFn {

		s.deployments.Items[0].Spec = expectedSpec
		s.deployments.Items[0].Labels = expectedLabels
		deploymentName := s.deployments.Items[0].GetName()

		r.log.Info(fmt.Sprintf("updating Deployment %s", deploymentName))

		r.err = r.client.Update(ctx, &s.deployments.Items[0])
		if r.err != nil {
			return nil
		}

		condition := serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonDeploymentUpdated,
			Message:            fmt.Sprintf("Deployment %s updated", deploymentName),
		}

		return buildStateFnUpdateStateFnFunctionCondition(condition)
	}
}

func stateFnUpdateDeploymentStatus(ctx context.Context, r *reconciler, s *systemState) stateFn {
	if r.err = ctx.Err(); r.err != nil {
		return nil
	}

	deploymentName := s.deployments.Items[0].GetName()

	// ready deployment
	if isDeploymentReady(s.deployments.Items[0]) {
		r.log.Info(fmt.Sprintf("deployment ready %q", deploymentName))

		condition := serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonDeploymentReady,
			Message:            fmt.Sprintf("Deployment %s is ready", deploymentName),
		}

		r.result = ctrl.Result{
			RequeueAfter: r.cfg.fn.FunctionReadyRequeueDuration,
		}

		return buildStateFnUpdateStateFnFunctionCondition(condition)
	}

	// unhealthy deployment
	if deploymentConditionFalseWithReason(s.deployments.Items[0], appsv1.DeploymentAvailable, MinimumReplicasUnavailable) {
		r.log.Info(fmt.Sprintf("deployment unhealthy: %q", deploymentName))

		condition := serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonMinReplicasNotAvailable,
			Message:            fmt.Sprintf("Minimum replcas not available for deployment %s", deploymentName),
		}

		return buildStateFnUpdateStateFnFunctionCondition(condition)
	}

	// deployment not ready
	if deploymentConditionTrue(s.deployments.Items[0], appsv1.DeploymentProgressing) {
		r.log.Info(fmt.Sprintf("deployment %q not ready", deploymentName))

		condition := serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonDeploymentWaiting,
			Message:            fmt.Sprintf("Deployment %s is not ready yet", deploymentName),
		}

		return buildStateFnUpdateStateFnFunctionCondition(condition)
	}

	// deployment failed
	r.log.Info(fmt.Sprintf("deployment %q failed", deploymentName))

	var yamlConditions []byte
	yamlConditions, r.err = yaml.Marshal(s.deployments.Items[0].Status.Conditions)

	if r.err != nil {
		return nil
	}

	msg := fmt.Sprintf("Deployment %s failed with condition: \n%s", deploymentName, yamlConditions)

	condition := serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonDeploymentFailed,
		Message:            msg,
	}

	return buildStateFnUpdateStateFnFunctionCondition(condition)
}

type buildDeploymentArgs struct {
	DockerPullAddress     string
	JaegerServiceEndpoint string
	PublisherProxyAddress string
	ImagePullAccountName  string
}

func buildFunctionDeployment(instance serverlessv1alpha1.Function, cfg buildDeploymentArgs) appsv1.Deployment {

	imageName := instance.BuildImageAddress(cfg.DockerPullAddress)
	deploymentLabels := instance.GetMergedLables()
	podLabels := instance.MergeLabels(instance.Spec.Labels, instance.DeploymentSelectorLabels())

	const volumeName = "tmp-dir"
	emptyDirVolumeSize := resource.MustParse("100Mi")

	rtmCfg := fnRuntime.GetRuntimeConfig(instance.Spec.Runtime)

	envs := append(instance.Spec.Env, rtmCfg.RuntimeEnvs...)

	deploymentEnvs := buildDeploymentEnvs(
		instance.GetNamespace(),
		cfg.JaegerServiceEndpoint,
		cfg.PublisherProxyAddress,
	)
	envs = append(envs, deploymentEnvs...)

	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
			Labels:       deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: instance.Spec.MinReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: instance.DeploymentSelectorLabels(), // this has to match spec.template.objectmeta.Labels
				// and also it has to be immutable
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels, // podLabels contains InternalFnLabels, so it's ok
					Annotations: map[string]string{
						"proxy.istio.io/config": "{ \"holdApplicationUntilProxyStarts\": true }",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: volumeName,
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium:    corev1.StorageMediumDefault,
								SizeLimit: &emptyDirVolumeSize,
							},
						},
					}},
					Containers: []corev1.Container{
						{
							Name:      functionContainerName,
							Image:     imageName,
							Env:       envs,
							Resources: instance.Spec.Resources,
							VolumeMounts: []corev1.VolumeMount{{
								Name: volumeName,
								/* needed in order to have python functions working:
								python functions need writable /tmp dir, but we disable writing to root filesystem via
								security context below. That's why we override this whole /tmp directory with emptyDir volume.
								We've decided to add this directory to be writable by all functions, as it may come in handy
								*/
								MountPath: "/tmp",
								ReadOnly:  false,
							}},
							/*
								In order to mark pod as ready we need to ensure the function is actually running and ready to serve traffic.
								We do this but first ensuring that sidecar is raedy by using "proxy.istio.io/config": "{ \"holdApplicationUntilProxyStarts\": true }", annotation
								Second thing is setting that probe which continously
							*/
							StartupProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: svcTargetPort,
									},
								},
								InitialDelaySeconds: 0,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								FailureThreshold:    30, // FailureThreshold * PeriodSeconds = 150s in this case, this should be enough for any function pod to start up
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: svcTargetPort,
									},
								},
								InitialDelaySeconds: 0, // startup probe exists, so delaying anything here doesn't make sense
								FailureThreshold:    1,
								PeriodSeconds:       5,
								TimeoutSeconds:      2,
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: svcTargetPort,
									},
								},
								FailureThreshold: 3,
								PeriodSeconds:    5,
								TimeoutSeconds:   4,
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add:  []corev1.Capability{},
									Drop: []corev1.Capability{"ALL"},
								},
								Privileged:               boolPtr(false),
								RunAsUser:                &functionUser,
								RunAsGroup:               &functionUser,
								RunAsNonRoot:             boolPtr(true),
								ReadOnlyRootFilesystem:   boolPtr(true),
								AllowPrivilegeEscalation: boolPtr(false),
							},
						},
					},
					ServiceAccountName: cfg.ImagePullAccountName,
				},
			},
		},
	}
}

func deploymentConditionTrueWithReason(deployment appsv1.Deployment, conditionType appsv1.DeploymentConditionType, reason string) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionTrue &&
				condition.Reason == reason
		}
	}
	return false
}

func isDeploymentReady(deployment appsv1.Deployment) bool {
	return deploymentConditionTrueWithReason(deployment, appsv1.DeploymentAvailable, MinimumReplicasAvailable) &&
		deploymentConditionTrueWithReason(deployment, appsv1.DeploymentProgressing, NewRSAvailableReason)
}

func deploymentConditionFalseWithReason(deployment appsv1.Deployment, conditionType appsv1.DeploymentConditionType, reason string) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionFalse &&
				condition.Reason == reason
		}
	}
	return false
}

func deploymentConditionTrue(deployment appsv1.Deployment, conditionType appsv1.DeploymentConditionType) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func functionDeploymentChanged(existing, expected appsv1.Deployment, scalingEnabled bool) bool {
	return !equalDeployments(existing, expected, scalingEnabled)
}

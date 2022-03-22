package serverless

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func calculateInlineImageTag(instance *serverlessv1alpha1.Function) string {
	hash := sha256.Sum256([]byte(strings.Join([]string{
		string(instance.GetUID()),
		instance.Spec.Source,
		instance.Spec.Deps,
		string(instance.Status.Runtime),
	}, "-")))

	return fmt.Sprintf("%x", hash)
}

func isOnConfigMapChange(instance *serverlessv1alpha1.Function, rtm runtime.Runtime, configMaps []corev1.ConfigMap, deployments []appsv1.Deployment, dockerConfig DockerConfig) bool {
	image := buildInlineImageAddress(instance, dockerConfig.PullAddress)
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

func buildJob(instance *serverlessv1alpha1.Function, rtmConfig runtime.Config, configMapName string, dockerConfig DockerConfig, config FunctionConfig) batchv1.Job {
	one := int32(1)
	zero := int32(0)
	rootUser := int64(0)
	optional := true

	imageName := buildInlineImageAddress(instance, dockerConfig.PushAddress)
	args := config.Build.ExecutorArgs
	args = append(args, fmt.Sprintf("%s=%s", destinationArg, imageName), fmt.Sprintf("--context=dir://%s", workspaceMountPath))

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-build-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
			Labels:       functionLabels(instance),
		},
		Spec: batchv1.JobSpec{
			Parallelism:           &one,
			Completions:           &one,
			ActiveDeadlineSeconds: nil,
			BackoffLimit:          &zero,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      functionLabels(instance),
					Annotations: istioSidecarInjectFalse,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "sources",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: configMapName},
								},
							},
						},
						{
							Name: "runtime",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: rtmConfig.DockerfileConfigMapName},
								},
							},
						},
						{
							Name: "credentials",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: dockerConfig.ActiveRegistryConfigSecretName,
									Items: []corev1.KeyToPath{
										{
											Key:  ".dockerconfigjson",
											Path: ".docker/config.json",
										},
									},
								},
							},
						},
						{
							Name: "registry-config",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: config.PackageRegistryConfigSecretName,
									Optional:   &optional,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "executor",
							Image:           config.Build.ExecutorImage,
							Args:            args,
							Resources:       instance.Spec.BuildResources,
							VolumeMounts:    getBuildJobVolumeMounts(rtmConfig),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "DOCKER_CONFIG", Value: "/docker/.docker/"},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: config.BuildServiceAccountName,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &rootUser,
					},
				},
			},
		},
	}
}

func isOnJobChange(instance *serverlessv1alpha1.Function, rtmCfg runtime.Config, jobs []batchv1.Job, deployments []appsv1.Deployment, gitOptions git.Options, dockerConfig DockerConfig, config FunctionConfig) bool {
	image := buildInlineImageAddress(instance, dockerConfig.PullAddress)
	buildStatus := getConditionStatus(instance.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)

	expectedJob := buildJob(instance, rtmCfg, "", dockerConfig, config)

	if len(deployments) == 1 &&
		deployments[0].Spec.Template.Spec.Containers[0].Image == image &&
		buildStatus != corev1.ConditionUnknown &&
		len(jobs) > 0 &&
		mapsEqual(expectedJob.GetLabels(), jobs[0].GetLabels()) {
		return buildStatus == corev1.ConditionFalse
	}

	return len(jobs) != 1 ||
		len(jobs[0].Spec.Template.Spec.Containers) != 1 ||
		// Compare image argument
		!equalJobs(jobs[0], expectedJob) ||
		!mapsEqual(expectedJob.GetLabels(), jobs[0].GetLabels()) ||
		buildStatus == corev1.ConditionUnknown ||
		buildStatus == corev1.ConditionFalse
}

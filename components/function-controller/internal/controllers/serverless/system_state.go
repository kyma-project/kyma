package serverless

import (
	"fmt"
	"path"
	"strings"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const DefaultDeploymentReplicas int32 = 1

type SystemState interface{}

// TODO extract interface
type systemState struct {
	instance    serverlessv1alpha2.Function
	image       string               // TODO make sure this is needed
	configMaps  corev1.ConfigMapList // TODO create issue to refactor this (only 1 config map should be here)
	deployments appsv1.DeploymentList
	jobs        batchv1.JobList
	services    corev1.ServiceList
	hpas        autoscalingv1.HorizontalPodAutoscalerList
}

var _ SystemState = systemState{}

func (s *systemState) internalFunctionLabels() map[string]string {
	labels := make(map[string]string, 3)

	labels[serverlessv1alpha2.FunctionNameLabel] = s.instance.Name
	labels[serverlessv1alpha2.FunctionManagedByLabel] = serverlessv1alpha2.FunctionControllerValue
	labels[serverlessv1alpha2.FunctionUUIDLabel] = string(s.instance.GetUID())

	return labels
}

func (s *systemState) functionLabels() map[string]string {
	internalLabels := s.internalFunctionLabels()
	functionLabels := s.instance.GetLabels()

	return mergeLabels(functionLabels, internalLabels)
}

func (s *systemState) buildImageAddress(registryAddress string) string {
	var imageTag string
	isGitType := s.instance.TypeOf(serverlessv1alpha2.FunctionTypeGit)
	if isGitType {
		imageTag = calculateGitImageTag(&s.instance)
	} else {
		imageTag = calculateInlineImageTag(&s.instance)
	}
	return fmt.Sprintf("%s/%s-%s:%s", registryAddress, s.instance.Namespace, s.instance.Name, imageTag)
}

// TODO to self - create issue to refactor this
func (s *systemState) inlineFnSrcChanged(dockerPullAddress string) bool {
	image := s.buildImageAddress(dockerPullAddress)
	configurationStatus := getConditionStatus(s.instance.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)
	rtm := fnRuntime.GetRuntime(s.instance.Spec.Runtime)
	labels := s.functionLabels()

	if len(s.deployments.Items) == 1 &&
		len(s.configMaps.Items) == 1 &&
		s.deployments.Items[0].Spec.Template.Spec.Containers[0].Image == image &&
		s.instance.Spec.Source.Inline.Source == s.configMaps.Items[0].Data[FunctionSourceKey] &&
		rtm.SanitizeDependencies(s.instance.Spec.Source.Inline.Dependencies) == s.configMaps.Items[0].Data[FunctionDepsKey] &&
		configurationStatus != corev1.ConditionUnknown &&
		mapsEqual(s.configMaps.Items[0].Labels, labels) {
		return false
	}

	return !(len(s.configMaps.Items) == 1 &&
		s.instance.Spec.Source.Inline.Source == s.configMaps.Items[0].Data[FunctionSourceKey] &&
		rtm.SanitizeDependencies(s.instance.Spec.Source.Inline.Dependencies) == s.configMaps.Items[0].Data[FunctionDepsKey] &&
		configurationStatus == corev1.ConditionTrue &&
		mapsEqual(s.configMaps.Items[0].Labels, labels))
}

func (s *systemState) buildConfigMap() corev1.ConfigMap {
	rtm := fnRuntime.GetRuntime(s.instance.Spec.Runtime)
	data := map[string]string{
		FunctionSourceKey: s.instance.Spec.Source.Inline.Source,
		FunctionDepsKey:   rtm.SanitizeDependencies(s.instance.Spec.Source.Inline.Dependencies),
	}
	labels := s.functionLabels()

	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       labels,
			GenerateName: fmt.Sprintf("%s-", s.instance.GetName()),
			Namespace:    s.instance.GetNamespace(),
		},
		Data: data,
	}
}

func (s *systemState) fnJobChanged(expectedJob batchv1.Job) bool {
	conditionStatus := getConditionStatus(
		s.instance.Status.Conditions,
		serverlessv1alpha2.ConditionBuildReady,
	)

	if len(s.deployments.Items) == 1 &&
		s.deployments.Items[0].Spec.Template.Spec.Containers[0].Image == s.image &&
		conditionStatus != corev1.ConditionUnknown &&
		len(s.jobs.Items) > 0 &&
		mapsEqual(expectedJob.GetLabels(), s.jobs.Items[0].GetLabels()) {

		return conditionStatus == corev1.ConditionFalse
	}

	return len(s.jobs.Items) != 1 ||
		len(s.jobs.Items[0].Spec.Template.Spec.Containers) != 1 ||
		// Compare image argument
		!equalJobs(s.jobs.Items[0], expectedJob) ||
		!mapsEqual(expectedJob.GetLabels(), s.jobs.Items[0].GetLabels()) ||
		conditionStatus == corev1.ConditionUnknown ||
		conditionStatus == corev1.ConditionFalse
}

var (
	one          = int32(1)
	zero         = int32(0)
	rootUser     = int64(0)
	functionUser = int64(1000)
	optionalTrue = true
)

func (s *systemState) buildGitJob(gitOptions git.Options, cfg cfg) batchv1.Job {
	imageName := s.buildImageAddress(cfg.docker.PushAddress)

	args := append(cfg.fn.Build.ExecutorArgs, fmt.Sprintf("%s=%s", destinationArg, imageName), fmt.Sprintf("--context=dir://%s", workspaceMountPath))
	if s.instance.Spec.RuntimeImageOverride != "" {
		args = append(args, fmt.Sprintf("--build-arg=base_image=%s", s.instance.Spec.RuntimeImageOverride))
	}
	rtmCfg := fnRuntime.GetRuntimeConfig(s.instance.Spec.Runtime)

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-build-", s.instance.GetName()),
			Namespace:    s.instance.GetNamespace(),
			Labels:       s.functionLabels(),
		},
		Spec: batchv1.JobSpec{
			Parallelism:           &one,
			Completions:           &one,
			ActiveDeadlineSeconds: nil,
			BackoffLimit:          &zero,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      s.functionLabels(),
					Annotations: istioSidecarInjectFalse,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "credentials",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: cfg.docker.ActiveRegistryConfigSecretName,
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
							Name: "runtime",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: rtmCfg.DockerfileConfigMapName},
								},
							},
						},
						{
							Name:         "workspace",
							VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
						},
						{
							Name: "registry-config",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: cfg.fn.PackageRegistryConfigSecretName,
									Optional:   &optionalTrue,
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "repo-fetcher",
							Image:           cfg.fn.Build.RepoFetcherImage,
							Env:             buildRepoFetcherEnvVars(&s.instance, gitOptions),
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workspace",
									MountPath: workspaceMountPath,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "executor",
							Image:           cfg.fn.Build.ExecutorImage,
							Args:            args,
							Resources:       s.instance.Spec.ResourceConfiguration.Build.Resources,
							VolumeMounts:    s.getGitBuildJobVolumeMounts(rtmCfg),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "DOCKER_CONFIG", Value: "/docker/.docker/"},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: cfg.fn.BuildServiceAccountName,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &rootUser,
					},
				},
			},
		},
	}
}

func (s *systemState) buildJob(configMapName string, cfg cfg) batchv1.Job {
	rtmCfg := fnRuntime.GetRuntimeConfig(s.instance.Spec.Runtime)
	imageName := s.buildImageAddress(cfg.docker.PushAddress)
	args := append(cfg.fn.Build.ExecutorArgs, fmt.Sprintf("%s=%s", destinationArg, imageName), fmt.Sprintf("--context=dir://%s", workspaceMountPath))
	if s.instance.Spec.RuntimeImageOverride != "" {
		args = append(args, fmt.Sprintf("--build-arg=base_image=%s", s.instance.Spec.RuntimeImageOverride))
	}
	labels := s.functionLabels()

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-build-", s.instance.GetName()),
			Namespace:    s.instance.GetNamespace(),
			Labels:       labels,
		},
		Spec: batchv1.JobSpec{
			Parallelism:           &one,
			Completions:           &one,
			ActiveDeadlineSeconds: nil,
			BackoffLimit:          &zero,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
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
									LocalObjectReference: corev1.LocalObjectReference{Name: rtmCfg.DockerfileConfigMapName},
								},
							},
						},
						{
							Name: "credentials",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: cfg.docker.ActiveRegistryConfigSecretName,
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
									SecretName: cfg.fn.PackageRegistryConfigSecretName,
									Optional:   &optionalTrue,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "executor",
							Image:           cfg.fn.Build.ExecutorImage,
							Args:            args,
							Resources:       s.instance.Spec.ResourceConfiguration.Build.Resources,
							VolumeMounts:    getBuildJobVolumeMounts(rtmCfg),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "DOCKER_CONFIG", Value: "/docker/.docker/"},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: cfg.fn.BuildServiceAccountName,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &rootUser,
					},
				},
			},
		},
	}
}

func (s *systemState) deploymentSelectorLabels() map[string]string {
	return mergeLabels(
		map[string]string{
			serverlessv1alpha2.FunctionResourceLabel: serverlessv1alpha2.FunctionResourceLabelDeploymentValue,
		},
		s.internalFunctionLabels(),
	)
}

func (s *systemState) podLabels() map[string]string {
	selectorLabels := s.deploymentSelectorLabels()
	return mergeLabels(s.instance.Spec.Template.Labels, selectorLabels)
}

type buildDeploymentArgs struct {
	DockerPullAddress     string
	JaegerServiceEndpoint string
	PublisherProxyAddress string
	ImagePullAccountName  string
}

func (s *systemState) buildDeployment(cfg buildDeploymentArgs) appsv1.Deployment {

	imageName := s.buildImageAddress(cfg.DockerPullAddress)
	deploymentLabels := s.functionLabels()
	podLabels := s.podLabels()

	const volumeName = "tmp-dir"
	emptyDirVolumeSize := resource.MustParse("100Mi")

	rtmCfg := fnRuntime.GetRuntimeConfig(s.instance.Spec.Runtime)

	envs := append(s.instance.Spec.Env, rtmCfg.RuntimeEnvs...)

	deploymentEnvs := buildDeploymentEnvs(
		s.instance.GetNamespace(),
		cfg.JaegerServiceEndpoint,
		cfg.PublisherProxyAddress,
	)
	envs = append(envs, deploymentEnvs...)

	minReplicas := DefaultDeploymentReplicas
	if s.instance.Spec.ScaleConfig != nil && s.instance.Spec.ScaleConfig.MinReplicas != nil {
		minReplicas = *s.instance.Spec.ScaleConfig.MinReplicas
	}

	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", s.instance.GetName()),
			Namespace:    s.instance.GetNamespace(),
			Labels:       deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: s.instance.Spec.ScaleConfig.MinReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: s.deploymentSelectorLabels(), // this has to match spec.template.objectmeta.Labels
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
							Resources: s.instance.Spec.ResourceConfiguration.Function.Resources,
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
								We do this but first ensuring that sidecar is ready by using "proxy.istio.io/config": "{ \"holdApplicationUntilProxyStarts\": true }", annotation
								Second thing is setting that probe which continuously
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

//TODO do not negate
func (s *systemState) deploymentEqual(d appsv1.Deployment) bool {
	return len(s.deployments.Items) == 1 &&
		equalDeployments(s.deployments.Items[0], d, isScalingEnabled(&s.instance))
}

func (s *systemState) hasDeploymentConditionTrueStatusWithReason(conditionType appsv1.DeploymentConditionType, reason string) bool {
	for _, condition := range s.deployments.Items[0].Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionTrue &&
				condition.Reason == reason
		}
	}
	return false
}

func (s *systemState) isDeploymentReady() bool {
	return s.hasDeploymentConditionTrueStatusWithReason(appsv1.DeploymentAvailable, MinimumReplicasAvailable) &&
		s.hasDeploymentConditionTrueStatusWithReason(appsv1.DeploymentProgressing, NewRSAvailableReason)
}

func (s *systemState) hasDeploymentConditionFalseStatusWithReason(conditionType appsv1.DeploymentConditionType, reason string) bool {
	for _, condition := range s.deployments.Items[0].Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionFalse &&
				condition.Reason == reason
		}
	}
	return false
}

func (s *systemState) hasDeploymentConditionTrueStatus(conditionType appsv1.DeploymentConditionType) bool {
	for _, condition := range s.deployments.Items[0].Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func (s *systemState) buildService() corev1.Service {
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.instance.GetName(),
			Namespace: s.instance.GetNamespace(),
			Labels:    s.functionLabels(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "http", // it has to be here for istio to work properly
				TargetPort: svcTargetPort,
				Port:       80,
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: s.deploymentSelectorLabels(),
		},
	}
}

func (s *systemState) svcChanged(expectedSvc corev1.Service) bool {
	return !equalServices(s.services.Items[0], expectedSvc)
}

func (s *systemState) hpaEqual(targetCPUUtilizationPercentage int32) bool {
	if len(s.deployments.Items) == 0 {
		return false
	}

	expected := s.buildHorizontalPodAutoscaler(targetCPUUtilizationPercentage)

	scalingEnabled := isScalingEnabled(&s.instance)

	numHpa := len(s.hpas.Items)
	return (scalingEnabled && numHpa != 1) ||
		(scalingEnabled && !s.equalHorizontalPodAutoscalers(expected)) ||
		(!scalingEnabled && numHpa != 0)
}

func (s *systemState) defaultReplicas() (int32, int32) {
	var min = int32(1)
	var max int32
	if s.instance.Spec.ScaleConfig == nil {
		return min, min
	}
	spec := s.instance.Spec
	if spec.ScaleConfig.MinReplicas != nil && *spec.ScaleConfig.MinReplicas > 0 {
		min = *spec.ScaleConfig.MinReplicas
	}
	// special case
	if spec.ScaleConfig.MaxReplicas == nil || min > *spec.ScaleConfig.MaxReplicas {
		max = min
	} else {
		max = *spec.ScaleConfig.MaxReplicas
	}
	return min, max
}

func (s *systemState) buildHorizontalPodAutoscaler(targetCPUUtilizationPercentage int32) autoscalingv1.HorizontalPodAutoscaler {
	minReplicas, maxReplicas := s.defaultReplicas()
	deploymentName := s.deployments.Items[0].GetName()
	return autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", s.instance.GetName()),
			Namespace:    s.instance.GetNamespace(),
			Labels:       s.functionLabels(),
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       deploymentName,
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
			MinReplicas:                    &minReplicas,
			MaxReplicas:                    maxReplicas,
			TargetCPUUtilizationPercentage: &targetCPUUtilizationPercentage,
		},
	}
}

func (s *systemState) equalHorizontalPodAutoscalers(expected autoscalingv1.HorizontalPodAutoscaler) bool {
	existing := s.hpas.Items[0]
	return equalInt32Pointer(existing.Spec.TargetCPUUtilizationPercentage, expected.Spec.TargetCPUUtilizationPercentage) &&
		equalInt32Pointer(existing.Spec.MinReplicas, expected.Spec.MinReplicas) &&
		existing.Spec.MaxReplicas == expected.Spec.MaxReplicas &&
		mapsEqual(existing.Labels, expected.Labels) &&
		existing.Spec.ScaleTargetRef.Name == expected.Spec.ScaleTargetRef.Name
}

func (s *systemState) gitFnSrcChanged(commit string) bool {
	return s.instance.Status.Commit == "" ||
		commit != s.instance.Status.Commit ||
		s.instance.Spec.Source.GitRepository.Reference != s.instance.Status.Reference ||
		s.instance.Spec.Runtime != s.instance.Status.Runtime ||
		s.instance.Spec.Source.GitRepository.BaseDir != s.instance.Status.BaseDir ||
		getConditionStatus(s.instance.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady) == corev1.ConditionFalse

}

func (s *systemState) getGitBuildJobVolumeMounts(rtmConfig runtime.Config) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
		// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
		// If COPY is not used, then the cache will not work
		{Name: "workspace", MountPath: path.Join(workspaceMountPath, "src"), SubPath: strings.TrimPrefix(s.instance.Spec.Source.GitRepository.BaseDir, "/")},
		{Name: "runtime", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "Dockerfile"), SubPath: "Dockerfile"},
	}
	// add package registry config volume mount depending on the used runtime
	volumeMounts = append(volumeMounts, getPackageConfigVolumeMountsForRuntime(rtmConfig.Runtime)...)
	return volumeMounts
}

func (s *systemState) jobFailed(p func(reason string) bool) bool {
	if len(s.jobs.Items) == 0 {
		return false
	}

	return jobFailed(s.jobs.Items[0], p)
}

package serverless

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"path"
	"strings"

	"k8s.io/utils/pointer"

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

const istioConfigLabelKey = "proxy.istio.io/config"
const istioEnableHoldUntilProxyStartLabelValue = "{ \"holdApplicationUntilProxyStarts\": true }"

type SystemState interface{}

// TODO extract interface
type systemState struct {
	instance    serverlessv1alpha2.Function
	fnImage     string               // TODO make sure this is needed
	configMaps  corev1.ConfigMapList // TODO create issue to refactor this (only 1 config map should be here)
	deployments appsv1.DeploymentList
	jobs        batchv1.JobList
	services    corev1.ServiceList
	hpas        autoscalingv1.HorizontalPodAutoscalerList
}

var _ SystemState = systemState{}

func internalFunctionLabels(fn serverlessv1alpha2.Function) map[string]string {
	intLabels := make(map[string]string, 3)

	intLabels[serverlessv1alpha2.FunctionNameLabel] = fn.Name
	intLabels[serverlessv1alpha2.FunctionManagedByLabel] = serverlessv1alpha2.FunctionControllerValue
	intLabels[serverlessv1alpha2.FunctionUUIDLabel] = string(fn.GetUID())

	return intLabels
}

func (s *systemState) functionLabels() map[string]string {
	internalLabels := internalFunctionLabels(s.instance)
	functionLabels := s.instance.GetLabels()

	return labels.Merge(functionLabels, internalLabels)
}

func (s *systemState) functionAnnotations() map[string]string {
	return map[string]string{
		"prometheus.io/port":   "80",
		"prometheus.io/path":   "/metrics",
		"prometheus.io/scrape": "true",
	}
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
	fnLabels := s.functionLabels()

	if len(s.deployments.Items) == 1 &&
		len(s.configMaps.Items) == 1 &&
		s.deployments.Items[0].Spec.Template.Spec.Containers[0].Image == image &&
		s.instance.Spec.Source.Inline.Source == s.configMaps.Items[0].Data[FunctionSourceKey] &&
		rtm.SanitizeDependencies(s.instance.Spec.Source.Inline.Dependencies) == s.configMaps.Items[0].Data[FunctionDepsKey] &&
		configurationStatus != corev1.ConditionUnknown &&
		mapsEqual(s.configMaps.Items[0].Labels, fnLabels) {
		return false
	}

	return !(len(s.configMaps.Items) == 1 &&
		s.instance.Spec.Source.Inline.Source == s.configMaps.Items[0].Data[FunctionSourceKey] &&
		rtm.SanitizeDependencies(s.instance.Spec.Source.Inline.Dependencies) == s.configMaps.Items[0].Data[FunctionDepsKey] &&
		configurationStatus == corev1.ConditionTrue &&
		mapsEqual(s.configMaps.Items[0].Labels, fnLabels))
}

func (s *systemState) buildConfigMap() corev1.ConfigMap {
	rtm := fnRuntime.GetRuntime(s.instance.Spec.Runtime)
	data := map[string]string{
		FunctionSourceKey: s.instance.Spec.Source.Inline.Source,
		FunctionDepsKey:   rtm.SanitizeDependencies(s.instance.Spec.Source.Inline.Dependencies),
	}
	fnLabels := s.functionLabels()

	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       fnLabels,
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
		s.deployments.Items[0].Spec.Template.Spec.Containers[0].Image == s.fnImage &&
		conditionStatus != corev1.ConditionUnknown &&
		len(s.jobs.Items) > 0 &&
		mapsEqual(expectedJob.GetLabels(), s.jobs.Items[0].GetLabels()) {

		return conditionStatus == corev1.ConditionFalse
	}

	return len(s.jobs.Items) != 1 ||
		len(s.jobs.Items[0].Spec.Template.Spec.Containers) != 1 ||
		// Compare fnImage argument
		!equalJobs(s.jobs.Items[0], expectedJob) ||
		!mapsEqual(expectedJob.GetLabels(), s.jobs.Items[0].GetLabels()) ||
		conditionStatus == corev1.ConditionUnknown ||
		conditionStatus == corev1.ConditionFalse
}

var (
	rootUser          = int64(0)
	rootUserGroup     = int64(0)
	functionUser      = int64(10001)
	functionUserGroup = int64(10001)
)

func (s *systemState) buildGitJob(gitOptions git.Options, cfg cfg) batchv1.Job {
	templateSpec := corev1.PodSpec{
		Volumes: []corev1.Volume{
			buildJobCredentialsVolume(cfg),
			buildRegistryConfigVolume(cfg),
			s.buildJobRuntimeVolume(),
			{
				Name:         "workspace",
				VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
			},
		},
		InitContainers: []corev1.Container{
			s.buildGitJobRepoFetcherContainer(gitOptions, cfg),
		},
		Containers: []corev1.Container{
			s.buildJobExecutorContainer(cfg, s.getGitBuildJobVolumeMounts()),
		},
		RestartPolicy: corev1.RestartPolicyNever,
	}
	enrichPodSpecWithSecurityContext(&templateSpec, rootUser, rootUserGroup)

	return s.buildJobJob(templateSpec)
}

func (s *systemState) buildJob(configMapName string, cfg cfg) batchv1.Job {
	templateSpec := corev1.PodSpec{
		Volumes: []corev1.Volume{
			buildJobCredentialsVolume(cfg),
			buildRegistryConfigVolume(cfg),
			s.buildJobRuntimeVolume(),
			{
				Name: "sources",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: configMapName},
					},
				},
			},
		},
		Containers: []corev1.Container{
			s.buildJobExecutorContainer(cfg, s.getBuildJobVolumeMounts()),
		},
		RestartPolicy: corev1.RestartPolicyNever,
	}
	enrichPodSpecWithSecurityContext(&templateSpec, rootUser, rootUserGroup)

	return s.buildJobJob(templateSpec)
}

func (s *systemState) buildJobJob(templateSpec corev1.PodSpec) batchv1.Job {
	fnLabels := s.functionLabels()
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-build-", s.instance.GetName()),
			Namespace:    s.instance.GetNamespace(),
			Labels:       fnLabels,
		},
		Spec: batchv1.JobSpec{
			Parallelism:           pointer.Int32(1),
			Completions:           pointer.Int32(1),
			ActiveDeadlineSeconds: nil,
			BackoffLimit:          pointer.Int32(0),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      fnLabels,
					Annotations: istioSidecarInjectFalse,
				},
				Spec: templateSpec,
			},
		},
	}
}

func (s *systemState) buildGitJobRepoFetcherContainer(gitOptions git.Options, cfg cfg) corev1.Container {
	return corev1.Container{
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
		SecurityContext: restrictiveContainerSecurityContext(),
	}
}

func (s *systemState) buildJobExecutorContainer(cfg cfg, volumeMounts []corev1.VolumeMount) corev1.Container {
	imageName := s.buildImageAddress(cfg.docker.PushAddress)
	args := append(cfg.fn.Build.ExecutorArgs,
		fmt.Sprintf("%s=%s", destinationArg, imageName),
		fmt.Sprintf("--context=dir://%s", workspaceMountPath))
	if s.instance.Spec.RuntimeImageOverride != "" {
		args = append(args,
			fmt.Sprintf("--build-arg=base_image=%s", s.instance.Spec.RuntimeImageOverride))
	}

	resourceRequirements := getBuildResourceRequirements(s.instance, cfg)

	return corev1.Container{
		Name:            "executor",
		Image:           cfg.fn.Build.ExecutorImage,
		Args:            args,
		Resources:       resourceRequirements,
		VolumeMounts:    volumeMounts,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Env: []corev1.EnvVar{
			{Name: "DOCKER_CONFIG", Value: "/docker/.docker/"},
		},
		SecurityContext: buildJobContainerSecurityContext(),
	}
}

func getBuildResourceRequirements(instance serverlessv1alpha2.Function, cfg cfg) corev1.ResourceRequirements {
	presets := cfg.fn.ResourceConfiguration.BuildJob.Resources.Presets.ToResourceRequirements()
	if instance.Spec.ResourceConfiguration != nil {
		return instance.Spec.ResourceConfiguration.Build.EffectiveResource(
			cfg.fn.ResourceConfiguration.BuildJob.Resources.DefaultPreset,
			presets)
	}
	return presets[cfg.fn.ResourceConfiguration.BuildJob.Resources.DefaultPreset]
}

func (s *systemState) getBuildJobVolumeMounts() []corev1.VolumeMount {
	rtmCfg := fnRuntime.GetRuntimeConfig(s.instance.Spec.Runtime)
	volumeMounts := []corev1.VolumeMount{
		// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
		// If COPY is not used, then the cache will not work
		{Name: "sources", ReadOnly: true, MountPath: path.Join(baseDir, rtmCfg.DependencyFile), SubPath: FunctionDepsKey},
		{Name: "sources", ReadOnly: true, MountPath: path.Join(baseDir, rtmCfg.FunctionFile), SubPath: FunctionSourceKey},
		{Name: "runtime", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "Dockerfile"), SubPath: "Dockerfile"},
		{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
	}
	// add package registry config volume mount depending on the used runtime
	volumeMounts = append(volumeMounts, getPackageConfigVolumeMountsForRuntime(rtmCfg.Runtime)...)
	return volumeMounts
}

func (s *systemState) getGitBuildJobVolumeMounts() []corev1.VolumeMount {
	rtmCfg := fnRuntime.GetRuntimeConfig(s.instance.Spec.Runtime)
	volumeMounts := []corev1.VolumeMount{
		{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
		// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
		// If COPY is not used, then the cache will not work
		{Name: "workspace", MountPath: path.Join(workspaceMountPath, "src"), SubPath: strings.TrimPrefix(s.instance.Spec.Source.GitRepository.BaseDir, "/")},
		{Name: "runtime", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "Dockerfile"), SubPath: "Dockerfile"},
	}
	// add package registry config volume mount depending on the used runtime
	volumeMounts = append(volumeMounts, getPackageConfigVolumeMountsForRuntime(rtmCfg.Runtime)...)
	return volumeMounts
}

func buildJobCredentialsVolume(cfg cfg) corev1.Volume {
	return corev1.Volume{
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
	}
}

func buildRegistryConfigVolume(cfg cfg) corev1.Volume {
	return corev1.Volume{
		Name: "registry-config",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: cfg.fn.PackageRegistryConfigSecretName,
				Optional:   pointer.Bool(true),
			},
		},
	}
}
func (s *systemState) buildJobRuntimeVolume() corev1.Volume {
	rtmCfg := fnRuntime.GetRuntimeConfig(s.instance.Spec.Runtime)
	return corev1.Volume{
		Name: "runtime",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: rtmCfg.DockerfileConfigMapName,
				},
			},
		},
	}
}

func (s *systemState) deploymentSelectorLabels() map[string]string {
	return labels.Merge(
		map[string]string{
			serverlessv1alpha2.FunctionResourceLabel: serverlessv1alpha2.FunctionResourceLabelDeploymentValue,
		},
		internalFunctionLabels(s.instance),
	)
}

func (s *systemState) podLabels() map[string]string {
	result := s.deploymentSelectorLabels()
	if s.instance.Spec.Template != nil && s.instance.Spec.Template.Labels != nil {
		result = labels.Merge(s.instance.Spec.Template.Labels, result)
	}
	if s.instance.Spec.Labels != nil {
		result = labels.Merge(s.instance.Spec.Labels, result)
	}
	return result
}

func (s *systemState) defaultAnnotations() map[string]string {
	return map[string]string{
		istioConfigLabelKey: istioEnableHoldUntilProxyStartLabelValue,
	}
}

func (s *systemState) podAnnotations() map[string]string {
	result := s.defaultAnnotations()
	if s.instance.Spec.Annotations != nil {
		result = labels.Merge(s.instance.Spec.Annotations, result)
	}
	result = labels.Merge(s.specialDeploymentAnnotations(), result)
	return result
}

func (s *systemState) specialDeploymentAnnotations() map[string]string {
	deployments := s.deployments.Items
	if len(deployments) == 0 {
		return map[string]string{}
	}
	deploymentAnnotations := deployments[0].Spec.Template.GetAnnotations()
	specialDeploymentAnnotations := map[string]string{}
	for _, k := range []string{
		"kubectl.kubernetes.io/restartedAt",
	} {
		if v, found := deploymentAnnotations[k]; found {
			specialDeploymentAnnotations[k] = v
		}
	}
	return specialDeploymentAnnotations
}

type buildDeploymentArgs struct {
	DockerPullAddress      string
	TraceCollectorEndpoint string
	PublisherProxyAddress  string
	ImagePullAccountName   string
}

func (s *systemState) buildDeployment(cfg buildDeploymentArgs, resourceConfig Resources) appsv1.Deployment {
	imageName := s.buildImageAddress(cfg.DockerPullAddress)

	const volumeName = "tmp-dir"
	emptyDirVolumeSize := resource.MustParse("100Mi")

	rtmCfg := fnRuntime.GetRuntimeConfig(s.instance.Spec.Runtime)

	envs := append(s.instance.Spec.Env, rtmCfg.RuntimeEnvs...)

	deploymentEnvs := buildDeploymentEnvs(
		s.instance.GetNamespace(),
		cfg.TraceCollectorEndpoint,
		cfg.PublisherProxyAddress,
	)
	envs = append(envs, deploymentEnvs...)

	secretVolumes, secretVolumeMounts := buildDeploymentSecretVolumes(s.instance.Spec.SecretMounts)

	volumes := []corev1.Volume{
		{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium:    corev1.StorageMediumDefault,
					SizeLimit: &emptyDirVolumeSize,
				},
			},
		},
	}
	volumes = append(volumes, secretVolumes...)

	volumeMounts := []corev1.VolumeMount{
		{
			Name: volumeName,
			/* needed in order to have python functions working:
			python functions need writable /tmp dir, but we disable writing to root filesystem via
			security context below. That's why we override this whole /tmp directory with emptyDir volume.
			We've decided to add this directory to be writable by all functions, as it may come in handy
			*/
			MountPath: "/tmp",
			ReadOnly:  false,
		},
	}
	volumeMounts = append(volumeMounts, secretVolumeMounts...)

	templateSpec := corev1.PodSpec{
		Volumes: volumes,
		Containers: []corev1.Container{
			{
				Name:         functionContainerName,
				Image:        imageName,
				Env:          envs,
				Resources:    getDeploymentResources(s.instance, resourceConfig),
				VolumeMounts: volumeMounts,
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
				SecurityContext: restrictiveContainerSecurityContext(),
			},
		},
		ServiceAccountName: cfg.ImagePullAccountName,
	}
	enrichPodSpecWithSecurityContext(&templateSpec, functionUser, functionUserGroup)

	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", s.instance.GetName()),
			Namespace:    s.instance.GetNamespace(),
			Labels:       s.functionLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: s.getReplicas(DefaultDeploymentReplicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: s.deploymentSelectorLabels(), // this has to match spec.template.objectmeta.Labels
				// and also it has to be immutable
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      s.podLabels(), // podLabels contains InternalFnLabels, so it's ok
					Annotations: s.podAnnotations(),
				},
				Spec: templateSpec,
			},
		},
	}
}

func getDeploymentResources(instance serverlessv1alpha2.Function, resourceCfg Resources) corev1.ResourceRequirements {
	presets := resourceCfg.Presets.ToResourceRequirements()
	if instance.Spec.ResourceConfiguration != nil {
		return instance.Spec.ResourceConfiguration.Build.EffectiveResource(
			resourceCfg.DefaultPreset,
			presets)
	}
	return presets[resourceCfg.DefaultPreset]
}

func buildDeploymentSecretVolumes(secretMounts []serverlessv1alpha2.SecretMount) (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {
	volumes = []corev1.Volume{}
	volumeMounts = []corev1.VolumeMount{}
	for _, secretMount := range secretMounts {
		volumeName := secretMount.SecretName

		volume := corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  secretMount.SecretName,
					DefaultMode: pointer.Int32(0666), //read and write only for everybody
					Optional:    pointer.Bool(false),
				},
			},
		}
		volumes = append(volumes, volume)

		volumeMount := corev1.VolumeMount{
			Name:      volumeName,
			ReadOnly:  true,
			MountPath: secretMount.MountPath,
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}
	return volumes, volumeMounts
}

func (s *systemState) getReplicas(defaultVal int32) *int32 {
	if s.instance.Spec.Replicas != nil {
		return s.instance.Spec.Replicas
	}
	return &defaultVal
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
			Name:        s.instance.GetName(),
			Namespace:   s.instance.GetNamespace(),
			Labels:      s.functionLabels(),
			Annotations: s.functionAnnotations(),
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
	return autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", s.instance.GetName()),
			Namespace:    s.instance.GetNamespace(),
			Labels:       s.functionLabels(),
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind:       serverlessv1alpha2.FunctionKind,
				Name:       s.instance.Name,
				APIVersion: serverlessv1alpha2.GroupVersion.String(),
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

func (s *systemState) jobFailed(p func(reason string) bool) bool {
	if len(s.jobs.Items) == 0 {
		return false
	}

	return jobFailed(s.jobs.Items[0], p)
}

// security context is set to fulfill the baseline security profile
// based on https://raw.githubusercontent.com/kyma-project/community/main/concepts/psp-replacement/baseline-pod-spec.yaml
func restrictiveContainerSecurityContext() *corev1.SecurityContext {
	defaultProcMount := corev1.DefaultProcMount
	return &corev1.SecurityContext{
		Privileged: pointer.Bool(false),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"ALL",
			},
		},
		ProcMount:              &defaultProcMount,
		ReadOnlyRootFilesystem: pointer.Bool(true),
	}
}

// build job requires additional permissions some than the restrictive version
func buildJobContainerSecurityContext() *corev1.SecurityContext {
	securityContext := restrictiveContainerSecurityContext()
	securityContext.Capabilities.Add = []corev1.Capability{
		"CHOWN",        // for "chown"
		"FOWNER",       // for "chmod"
		"SETGID",       // for "fork"
		"DAC_OVERRIDE", // for "open"
	}
	securityContext.ReadOnlyRootFilesystem = pointer.Bool(false)
	return securityContext
}

// security context is set to fulfill the baseline security profile
// based on https://raw.githubusercontent.com/kyma-project/community/main/concepts/psp-replacement/baseline-pod-spec.yaml
func enrichPodSpecWithSecurityContext(ps *corev1.PodSpec, user int64, userGroup int64) {
	ps.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser:  &user,
		RunAsGroup: &userGroup,
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
	ps.HostNetwork = false
	ps.HostPID = false
	ps.HostIPC = false
}

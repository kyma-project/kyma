package serverless

import (
	"fmt"
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kyma-project/kyma/components/function-controller/internal/git"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const (
	destinationArg        = "--destination"
	functionContainerName = "function"
	baseDir               = "/workspace/src/"
	workspaceMountPath    = "/workspace"
)

var (
	istioSidecarInjectFalse = map[string]string{
		"sidecar.istio.io/inject": "false",
	}
	svcTargetPort = intstr.FromInt(8080) // https://github.com/kubeless/runtimes/blob/master/stable/nodejs/kubeless.js#L28
)

func boolPtr(b bool) *bool {
	return &b
}

func (r *FunctionReconciler) buildConfigMap(instance *serverlessv1alpha1.Function, rtm runtime.Runtime) corev1.ConfigMap {
	data := map[string]string{
		FunctionSourceKey: instance.Spec.Source,
		FunctionDepsKey:   rtm.SanitizeDependencies(instance.Spec.Deps),
	}

	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       r.functionLabels(instance),
			GenerateName: fmt.Sprintf("%s-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
		},
		Data: data,
	}
}

func (r *FunctionReconciler) buildJob(instance *serverlessv1alpha1.Function, rtmConfig runtime.Config, configMapName string, dockerConfig DockerConfig) batchv1.Job {
	one := int32(1)
	zero := int32(0)
	rootUser := int64(0)
	optional := true

	imageName := r.buildImageAddress(instance, dockerConfig.PushAddress)
	args := r.config.Build.ExecutorArgs
	args = append(args, fmt.Sprintf("%s=%s", destinationArg, imageName), fmt.Sprintf("--context=dir://%s", workspaceMountPath))

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-build-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
			Labels:       r.functionLabels(instance),
		},
		Spec: batchv1.JobSpec{
			Parallelism:           &one,
			Completions:           &one,
			ActiveDeadlineSeconds: nil,
			BackoffLimit:          &zero,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.functionLabels(instance),
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
									SecretName: r.config.PackageRegistryConfigSecretName,
									Optional:   &optional,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "executor",
							Image:           r.config.Build.ExecutorImage,
							Args:            args,
							Resources:       instance.Spec.BuildResources,
							VolumeMounts:    r.getBuildJobVolumeMounts(rtmConfig),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "DOCKER_CONFIG", Value: "/docker/.docker/"},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: r.config.BuildServiceAccountName,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &rootUser,
					},
				},
			},
		},
	}
}

func (r *FunctionReconciler) getBuildJobVolumeMounts(rtmConfig runtime.Config) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
		// If COPY is not used, then the cache will not work
		{Name: "sources", ReadOnly: true, MountPath: path.Join(baseDir, rtmConfig.DependencyFile), SubPath: FunctionDepsKey},
		{Name: "sources", ReadOnly: true, MountPath: path.Join(baseDir, rtmConfig.FunctionFile), SubPath: FunctionSourceKey},
		{Name: "runtime", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "Dockerfile"), SubPath: "Dockerfile"},
		{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
	}
	// add package registry config volume mount depending on the used runtime
	volumeMounts = append(volumeMounts, r.getPackageConfigVolumeMountsForRuntime(rtmConfig.Runtime)...)
	return volumeMounts
}

func buildRepoFetcherEnvVars(instance *serverlessv1alpha1.Function, gitOptions git.Options) []corev1.EnvVar {
	vars := []corev1.EnvVar{
		{
			Name:  "APP_REPOSITORY_URL",
			Value: gitOptions.URL,
		},
		{
			Name:  "APP_REPOSITORY_COMMIT",
			Value: instance.Status.Repository.Reference,
		},
		{
			Name:  "APP_MOUNT_PATH",
			Value: workspaceMountPath,
		},
	}

	if gitOptions.Auth != nil {
		vars = append(vars, corev1.EnvVar{
			Name:  "APP_REPOSITORY_AUTH_TYPE",
			Value: string(gitOptions.Auth.Type),
		})

		switch gitOptions.Auth.Type {
		case git.RepositoryAuthBasic:
			vars = append(vars, []corev1.EnvVar{
				{
					Name: "APP_REPOSITORY_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: gitOptions.Auth.SecretName,
							},
							Key: git.UsernameKey,
						},
					},
				},
				{
					Name: "APP_REPOSITORY_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: gitOptions.Auth.SecretName,
							},
							Key: git.PasswordKey,
						},
					},
				},
			}...)
			break
		case git.RepositoryAuthSSHKey:
			vars = append(vars, corev1.EnvVar{
				Name: "APP_REPOSITORY_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: gitOptions.Auth.SecretName,
						},
						Key: git.KeyKey,
					},
				},
			})
			if _, ok := gitOptions.Auth.Credentials[git.PasswordKey]; ok {
				vars = append(vars, corev1.EnvVar{
					Name: "APP_REPOSITORY_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: gitOptions.Auth.SecretName,
							},
							Key: git.PasswordKey,
						},
					},
				})
			}
			break
		}
	}

	return vars
}

func (r *FunctionReconciler) buildGitJob(instance *serverlessv1alpha1.Function, gitOptions git.Options, rtmConfig runtime.Config, dockerConfig DockerConfig) batchv1.Job {
	imageName := r.buildImageAddress(instance, dockerConfig.PushAddress)
	args := r.config.Build.ExecutorArgs
	args = append(args, fmt.Sprintf("%s=%s", destinationArg, imageName), fmt.Sprintf("--context=dir://%s", workspaceMountPath))

	one := int32(1)
	zero := int32(0)
	rootUser := int64(0)
	optional := true

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-build-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
			Labels:       r.functionLabels(instance),
		},
		Spec: batchv1.JobSpec{
			Parallelism:           &one,
			Completions:           &one,
			ActiveDeadlineSeconds: nil,
			BackoffLimit:          &zero,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.functionLabels(instance),
					Annotations: istioSidecarInjectFalse,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
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
							Name: "runtime",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: rtmConfig.DockerfileConfigMapName},
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
									SecretName: r.config.PackageRegistryConfigSecretName,
									Optional:   &optional,
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "repo-fetcher",
							Image:           r.config.Build.RepoFetcherImage,
							Env:             buildRepoFetcherEnvVars(instance, gitOptions),
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
							Image:           r.config.Build.ExecutorImage,
							Args:            args,
							Resources:       instance.Spec.BuildResources,
							VolumeMounts:    r.getGitBuildJobVolumeMounts(instance, rtmConfig),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "DOCKER_CONFIG", Value: "/docker/.docker/"},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: r.config.BuildServiceAccountName,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &rootUser,
					},
				},
			},
		},
	}
}

func (r *FunctionReconciler) getGitBuildJobVolumeMounts(instance *serverlessv1alpha1.Function, rtmConfig runtime.Config) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
		// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
		// If COPY is not used, then the cache will not work
		{Name: "workspace", MountPath: path.Join(workspaceMountPath, "src"), SubPath: strings.TrimPrefix(instance.Spec.BaseDir, "/")},
		{Name: "runtime", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "Dockerfile"), SubPath: "Dockerfile"},
	}
	// add package registry config volume mount depending on the used runtime
	volumeMounts = append(volumeMounts, r.getPackageConfigVolumeMountsForRuntime(rtmConfig.Runtime)...)
	return volumeMounts

}

func (r *FunctionReconciler) buildDeployment(instance *serverlessv1alpha1.Function, rtmConfig runtime.Config, dockerConfig DockerConfig) appsv1.Deployment {
	imageName := r.buildImageAddress(instance, dockerConfig.PullAddress)
	deploymentLabels := r.functionLabels(instance)
	podLabels := r.podLabels(instance)

	functionUser := int64(1000)
	const volumeName = "tmp-dir"
	emptyDirVolumeSize := resource.MustParse("100Mi")

	envs := append(instance.Spec.Env, rtmConfig.RuntimeEnvs...)
	envs = append(envs, envVarsForDeployment...)

	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
			Labels:       deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: instance.Spec.MinReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.deploymentSelectorLabels(instance), // this has to match spec.template.objectmeta.Labels
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
								PeriodSeconds:       1, // the lowest acceptable value, we should check it even more often but k8s doesn't let us
								SuccessThreshold:    1,
								FailureThreshold:    120, // FailureThreshold * PeriodSeconds = 120s in this case, this should be enough for any function pod to start up
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
					ServiceAccountName: r.config.ImagePullAccountName,
				},
			},
		},
	}
}

func (r *FunctionReconciler) buildService(instance *serverlessv1alpha1.Function) corev1.Service {
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName(),
			Namespace: instance.GetNamespace(),
			Labels:    r.functionLabels(instance),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "http", // it has to be here for istio to work properly
				TargetPort: svcTargetPort,
				Port:       80,
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: r.deploymentSelectorLabels(instance),
		},
	}
}

func (r *FunctionReconciler) buildHorizontalPodAutoscaler(instance *serverlessv1alpha1.Function, deploymentName string) autoscalingv1.HorizontalPodAutoscaler {
	minReplicas, maxReplicas := r.defaultReplicas(instance.Spec)
	return autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
			Labels:       r.functionLabels(instance),
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       deploymentName,
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
			MinReplicas:                    &minReplicas,
			MaxReplicas:                    maxReplicas,
			TargetCPUUtilizationPercentage: &r.config.TargetCPUUtilizationPercentage,
		},
	}
}

func (r *FunctionReconciler) defaultReplicas(spec serverlessv1alpha1.FunctionSpec) (int32, int32) {
	min, max := int32(1), int32(1)
	if spec.MinReplicas != nil && *spec.MinReplicas > 0 {
		min = *spec.MinReplicas
	}
	// special case
	if spec.MaxReplicas == nil || min > *spec.MaxReplicas {
		max = min
	} else {
		max = *spec.MaxReplicas
	}
	return min, max
}

func (r *FunctionReconciler) buildImageAddress(instance *serverlessv1alpha1.Function, registryAddress string) string {
	var imageTag string
	if instance.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		imageTag = r.calculateGitImageTag(instance)
	} else {
		imageTag = r.calculateImageTag(instance)
	}
	return fmt.Sprintf("%s/%s-%s:%s", registryAddress, instance.Namespace, instance.Name, imageTag)
}

func (r *FunctionReconciler) functionLabels(instance *serverlessv1alpha1.Function) map[string]string {
	return r.mergeLabels(instance.GetLabels(), r.internalFunctionLabels(instance))
}

func (r *FunctionReconciler) internalFunctionLabels(instance *serverlessv1alpha1.Function) map[string]string {
	labels := make(map[string]string, 3)

	labels[serverlessv1alpha1.FunctionNameLabel] = instance.Name
	labels[serverlessv1alpha1.FunctionManagedByLabel] = serverlessv1alpha1.FunctionControllerValue
	labels[serverlessv1alpha1.FunctionUUIDLabel] = string(instance.GetUID())

	return labels
}

func (r *FunctionReconciler) deploymentSelectorLabels(instance *serverlessv1alpha1.Function) map[string]string {
	return r.mergeLabels(map[string]string{serverlessv1alpha1.FunctionResourceLabel: serverlessv1alpha1.FunctionResourceLabelDeploymentValue}, r.internalFunctionLabels(instance))
}

func (r *FunctionReconciler) podLabels(instance *serverlessv1alpha1.Function) map[string]string {
	return r.mergeLabels(instance.Spec.Labels, r.deploymentSelectorLabels(instance))
}

func (r *FunctionReconciler) mergeLabels(labelsCollection ...map[string]string) map[string]string {
	result := make(map[string]string, 0)
	for _, labels := range labelsCollection {
		for key, value := range labels {
			result[key] = value
		}
	}
	return result
}

func (r *FunctionReconciler) getPackageConfigVolumeMountsForRuntime(rtm serverlessv1alpha1.Runtime) []corev1.VolumeMount {
	switch rtm {
	case serverlessv1alpha1.Nodejs12, serverlessv1alpha1.Nodejs14:
		return []corev1.VolumeMount{{Name: "registry-config", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "registry-config/.npmrc"), SubPath: ".npmrc"}}
	case serverlessv1alpha1.Python38:
		return []corev1.VolumeMount{{Name: "registry-config", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "registry-config/pip.conf"), SubPath: "pip.conf"}}
	}
	return nil
}

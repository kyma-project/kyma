package serverless

import (
	"fmt"
	"path"
	"strings"

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

var istioSidecarInjectFalse = map[string]string{
	"sidecar.istio.io/inject": "false",
}

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

func (r *FunctionReconciler) buildJob(instance *serverlessv1alpha1.Function, rtmConfig runtime.Config, configMapName string) batchv1.Job {
	one := int32(1)
	zero := int32(0)
	rootUser := int64(0)

	imageName := r.buildImageAddressForPush(instance)
	args := r.config.Build.ExecutorArgs
	args = append(args, fmt.Sprintf("%s=%s", destinationArg, imageName), fmt.Sprintf("--context=dir://%s", workspaceMountPath))

	r.Log.Info(r.config.BuildServiceAccountName)

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
									SecretName: r.config.ImageRegistryDockerConfigSecretName,
									Items: []corev1.KeyToPath{
										{
											Key:  ".dockerconfigjson",
											Path: ".docker/config.json",
										},
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:      "executor",
							Image:     r.config.Build.ExecutorImage,
							Args:      args,
							Resources: instance.Spec.BuildResources,
							VolumeMounts: []corev1.VolumeMount{
								// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
								// If COPY is not used, then the cache will not work
								{Name: "sources", ReadOnly: true, MountPath: path.Join(baseDir, rtmConfig.DependencyFile), SubPath: FunctionDepsKey},
								{Name: "sources", ReadOnly: true, MountPath: path.Join(baseDir, rtmConfig.FunctionFile), SubPath: FunctionSourceKey},
								{Name: "runtime", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "Dockerfile"), SubPath: "Dockerfile"},
								{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
							},
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

func (r *FunctionReconciler) buildGitJob(instance *serverlessv1alpha1.Function, gitOptions git.Options, rtmConfig runtime.Config) batchv1.Job {
	imageName := r.buildImageAddressForPush(instance)
	args := r.config.Build.ExecutorArgs
	args = append(args, fmt.Sprintf("%s=%s", destinationArg, imageName), fmt.Sprintf("--context=dir://%s", workspaceMountPath))

	one := int32(1)
	zero := int32(0)
	rootUser := int64(0)

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
									SecretName: r.config.ImageRegistryDockerConfigSecretName,
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
							Name:      "executor",
							Image:     r.config.Build.ExecutorImage,
							Args:      args,
							Resources: instance.Spec.BuildResources,
							VolumeMounts: []corev1.VolumeMount{
								{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
								// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
								// If COPY is not used, then the cache will not work
								{Name: "workspace", MountPath: path.Join(workspaceMountPath, "src"), SubPath: strings.TrimPrefix(instance.Spec.BaseDir, "/")},
								{Name: "runtime", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "Dockerfile"), SubPath: "Dockerfile"},
							},
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

func (r *FunctionReconciler) buildDeployment(instance *serverlessv1alpha1.Function, rtmConfig runtime.Config) appsv1.Deployment {
	imageName := r.buildImageAddress(instance)
	deploymentLabels := r.functionLabels(instance)
	podLabels := r.podLabels(instance)

	functionUser := int64(1000)

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
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            functionContainerName,
							Image:           imageName,
							Env:             envs,
							Resources:       instance.Spec.Resources,
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
				Name:       "http",               // it has to be here for istio to work properly
				TargetPort: intstr.FromInt(8080), // https://github.com/kubeless/runtimes/blob/master/stable/nodejs/kubeless.js#L28
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

func (r *FunctionReconciler) buildImageAddressForPush(instance *serverlessv1alpha1.Function) string {
	if r.config.Docker.InternalRegistryEnabled {
		return r.buildInternalImageAddress(instance)
	}
	return r.buildImageAddress(instance)
}

func (r *FunctionReconciler) buildInternalImageAddress(instance *serverlessv1alpha1.Function) string {
	var imageTag string
	if instance.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		imageTag = r.calculateGitImageTag(instance)
	} else {
		imageTag = r.calculateImageTag(instance)
	}

	return fmt.Sprintf("%s/%s-%s:%s", r.config.Docker.InternalServerAddress, instance.Namespace, instance.Name, imageTag)
}

func (r *FunctionReconciler) buildImageAddress(instance *serverlessv1alpha1.Function) string {
	var imageTag string
	if instance.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		imageTag = r.calculateGitImageTag(instance)
	} else {
		imageTag = r.calculateImageTag(instance)
	}
	return fmt.Sprintf("%s/%s-%s:%s", r.config.Docker.RegistryAddress, instance.Namespace, instance.Name, imageTag)
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

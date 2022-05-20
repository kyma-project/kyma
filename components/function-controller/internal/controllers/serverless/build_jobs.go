package serverless

import (
	"fmt"
	"path"
	"strings"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	one          = int32(1)
	zero         = int32(0)
	rootUser     = int64(0)
	functionUser = int64(1000)
	optionalTrue = true
)

func gitBuildVolumeMounts(rtmConfig runtime.Config, baseDir string) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
		// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
		// If COPY is not used, then the cache will not work
		{Name: "workspace", MountPath: path.Join(workspaceMountPath, "src"), SubPath: strings.TrimPrefix(baseDir, "/")},
		{Name: "runtime", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "Dockerfile"), SubPath: "Dockerfile"},
	}
	// add package registry config volume mount depending on the used runtime
	volumeMounts = append(volumeMounts, getPackageConfigVolumeMountsForRuntime(rtmConfig.Runtime)...)
	return volumeMounts
}

func buildGitJob(instance serverlessv1alpha1.Function, gitOptions git.Options, cfg cfg) batchv1.Job {
	imageName := instance.BuildImageAddress(cfg.docker.PushAddress)

	args := append(cfg.fn.Build.ExecutorArgs, fmt.Sprintf("%s=%s", destinationArg, imageName), fmt.Sprintf("--context=dir://%s", workspaceMountPath))
	if instance.Spec.RuntimeImageOverride != "" {
		args = append(args, fmt.Sprintf("--build-arg=base_image=%s", instance.Spec.RuntimeImageOverride))
	}
	rtmCfg := fnRuntime.GetRuntimeConfig(instance.Spec.Runtime)

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-build-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
			Labels:       instance.GetMergedLables(),
		},
		Spec: batchv1.JobSpec{
			Parallelism:           &one,
			Completions:           &one,
			ActiveDeadlineSeconds: nil,
			BackoffLimit:          &zero,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      instance.GetMergedLables(),
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
							Env:             buildRepoFetcherEnvVars(&instance, gitOptions),
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
							Resources:       instance.Spec.BuildResources,
							VolumeMounts:    gitBuildVolumeMounts(rtmCfg, instance.Spec.BaseDir),
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

func buildJob(instance serverlessv1alpha1.Function, configMapName string, cfg cfg) batchv1.Job {
	rtmCfg := fnRuntime.GetRuntimeConfig(instance.Spec.Runtime)
	imageName := instance.BuildImageAddress(cfg.docker.PushAddress)
	args := append(cfg.fn.Build.ExecutorArgs, fmt.Sprintf("%s=%s", destinationArg, imageName), fmt.Sprintf("--context=dir://%s", workspaceMountPath))
	if instance.Spec.RuntimeImageOverride != "" {
		args = append(args, fmt.Sprintf("--build-arg=base_image=%s", instance.Spec.RuntimeImageOverride))
	}
	labels := instance.GetMergedLables()

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-build-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
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
							Resources:       instance.Spec.BuildResources,
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

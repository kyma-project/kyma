package serverless

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const (
	destinationArg        = "--destination"
	functionContainerName = "lambda"
)

func (r *FunctionReconciler) buildConfigMap(instance *serverlessv1alpha1.Function) corev1.ConfigMap {
	data := map[string]string{
		configMapHandler:  configMapHandler,
		configMapFunction: instance.Spec.Source,
		configMapDeps:     r.sanitizeDependencies(instance.Spec.Deps),
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

func (r *FunctionReconciler) buildJob(instance *serverlessv1alpha1.Function, configMapName string) batchv1.Job {
	one := int32(1)
	zero := int32(0)

	imageName := r.buildImageAddressForPush(instance)
	args := r.config.Build.ExecutorArgs
	args = append(args, fmt.Sprintf("%s=%s", destinationArg, imageName), "--context=dir:///workspace")

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
					Labels: r.functionLabels(instance),
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
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
									LocalObjectReference: corev1.LocalObjectReference{Name: r.config.Build.RuntimeConfigMapName},
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
							Name:  "executor",
							Image: r.config.Build.ExecutorImage,
							Args:  args,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: r.config.Build.LimitsMemoryValue,
									corev1.ResourceCPU:    r.config.Build.LimitsCPUValue,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: r.config.Build.RequestsMemoryValue,
									corev1.ResourceCPU:    r.config.Build.RequestsCPUValue,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
								// If COPY is not used, then the cache will not work
								{Name: "sources", ReadOnly: true, MountPath: "/workspace/src/package.json", SubPath: "package.json"},
								{Name: "sources", ReadOnly: true, MountPath: "/workspace/src/handler.js", SubPath: "handler.js"},
								{Name: "runtime", ReadOnly: true, MountPath: "/workspace/Dockerfile", SubPath: "Dockerfile"},
								{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "DOCKER_CONFIG", Value: "/docker/.docker/"},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: r.config.ImagePullAccountName,
				},
			},
		},
	}
}

func (r *FunctionReconciler) buildDeployment(instance *serverlessv1alpha1.Function) appsv1.Deployment {
	imageName := r.buildImageAddress(instance)
	deploymentLabels := r.functionLabels(instance)
	podLabels := r.podLabels(instance)

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
							Env:             append(instance.Spec.Env, envVarsForDeployment...),
							Resources:       instance.Spec.Resources,
							ImagePullPolicy: corev1.PullIfNotPresent,
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
	imageTag := r.calculateImageTag(instance)
	return fmt.Sprintf("%s/%s-%s:%s", r.config.Docker.InternalServerAddress, instance.Namespace, instance.Name, imageTag)
}

func (r *FunctionReconciler) buildImageAddress(instance *serverlessv1alpha1.Function) string {
	imageTag := r.calculateImageTag(instance)
	return fmt.Sprintf("%s/%s-%s:%s", r.config.Docker.RegistryAddress, instance.Namespace, instance.Name, imageTag)
}

func (r *FunctionReconciler) sanitizeDependencies(dependencies string) string {
	result := "{}"
	if strings.Trim(dependencies, " ") != "" {
		result = dependencies
	}

	return result
}

func (r *FunctionReconciler) functionLabels(instance *serverlessv1alpha1.Function) map[string]string {
	return r.mergeLabels(instance.GetLabels(), r.internalFunctionLabels(instance))
}

func (r *FunctionReconciler) internalFunctionLabels(instance *serverlessv1alpha1.Function) map[string]string {
	labels := make(map[string]string, 3)

	labels[serverlessv1alpha1.FunctionNameLabel] = instance.Name
	labels[serverlessv1alpha1.FunctionManagedByLabel] = "function-controller"
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

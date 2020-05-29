package serverless

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const (
	destinationArg        = "--destination"
	functionContainerName = "function"
)

func (r *FunctionReconciler) buildConfigMap(instance *serverlessv1alpha1.Function) corev1.ConfigMap {
	data := map[string]string{
		configMapHandler:  configMapHandler,
		configMapFunction: instance.Spec.Source,
		configMapDeps:     r.sanitizeDependencies(instance.Spec.Deps),
	}
	labels := r.functionLabels(instance)

	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       labels,
			GenerateName: fmt.Sprintf("%s-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
		},
		Data: data,
	}
}

func (r *FunctionReconciler) buildJob(instance *serverlessv1alpha1.Function, configMapName string) batchv1.Job {
	imageName := r.buildInternalImageAddress(instance)
	one := int32(1)
	zero := int32(0)

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
								Secret: &corev1.SecretVolumeSource{SecretName: r.config.ImagePullSecretName},
							},
						},
						{
							Name:         "tekton-home",
							VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
						},
						{
							Name:         "tekton-workspace",
							VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:    "credential-initializer",
							Image:   r.config.Build.CredsInitImage,
							Command: []string{"/ko-app/creds-init"},
							Args:    []string{fmt.Sprintf("-basic-docker=credentials=http://%s", imageName)},
							Env: []corev1.EnvVar{
								{Name: "HOME", Value: "/tekton/home"},
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "tekton-home", ReadOnly: false, MountPath: "/tekton/home"},
								{Name: "credentials", ReadOnly: false, MountPath: "/tekton/creds-secrets/credentials"},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "executor",
							Image: r.config.Build.ExecutorImage,
							Args:  append(r.config.Build.ExecutorArgs, fmt.Sprintf("%s=%s", destinationArg, imageName), "--context=dir:///workspace"),
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
								{Name: "tekton-home", ReadOnly: false, MountPath: "/tekton/home"},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "DOCKER_CONFIG", Value: "/tekton/home/.docker/"},
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
	imageName := r.buildExternalImageAddress(instance)
	deploymentLabels := r.functionLabels(instance)

	podLabels := r.servicePodLabels(instance)

	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
			Labels:       deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: instance.Spec.MinReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.internalFunctionLabels(instance), // this has to be same as spec.template.objectmeta.Lables
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
			Selector: r.internalFunctionLabels(instance),
		},
	}
}

func (r *FunctionReconciler) buildInternalImageAddress(instance *serverlessv1alpha1.Function) string {
	imageTag := r.calculateImageTag(instance)
	return fmt.Sprintf("%s/%s-%s:%s", r.config.Docker.Address, instance.Namespace, instance.Name, imageTag)
}

func (r *FunctionReconciler) buildExternalImageAddress(instance *serverlessv1alpha1.Function) string {
	imageTag := r.calculateImageTag(instance)
	return fmt.Sprintf("%s/%s-%s:%s", r.config.Docker.ExternalAddress, instance.Namespace, instance.Name, imageTag)
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

func (r *FunctionReconciler) servicePodLabels(instance *serverlessv1alpha1.Function) map[string]string {
	return r.mergeLabels(instance.Spec.Labels, r.internalFunctionLabels(instance))
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

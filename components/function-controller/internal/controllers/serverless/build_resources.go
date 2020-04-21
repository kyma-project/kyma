package serverless

import (
	"fmt"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func (r *FunctionReconciler) buildJob(instance *serverlessv1alpha1.Function, configMapName string) batchv1.Job {
	imageName := r.buildInternalImageAddress(instance)
	one := int32(1)
	zero := int32(0)

	job := batchv1.Job{
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
							Args:  []string{fmt.Sprintf("--destination=%s", imageName), "--insecure", "--skip-tls-verify"},
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
								{Name: "sources", ReadOnly: true, MountPath: "/src"},
								{Name: "runtime", ReadOnly: true, MountPath: "/workspace"},
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

	return job
}

func (r *FunctionReconciler) buildService(log logr.Logger, instance *serverlessv1alpha1.Function, oldService *servingv1.Service) servingv1.Service {
	imageName := r.buildExternalImageAddress(instance)
	annotations := map[string]string{
		"autoscaling.knative.dev/minScale": "1",
		"autoscaling.knative.dev/maxScale": "1",
	}
	if instance.Spec.MinReplicas != nil {
		annotations["autoscaling.knative.dev/minScale"] = fmt.Sprintf("%d", *instance.Spec.MinReplicas)
	}
	if instance.Spec.MaxReplicas != nil {
		annotations["autoscaling.knative.dev/maxScale"] = fmt.Sprintf("%d", *instance.Spec.MaxReplicas)
	}
	serviceLabels := r.functionLabels(instance)
	serviceLabels["serving.knative.dev/visibility"] = "cluster-local"

	bindingAnnotation := ""
	if oldService != nil {
		bindingAnnotation = oldService.GetAnnotations()[serviceBindingUsagesAnnotation]
	}
	podLabels := r.servingPodLabels(log, instance, bindingAnnotation)

	service := servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName(),
			Namespace: instance.GetNamespace(),
			Labels:    serviceLabels,
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: annotations,
						Labels:      podLabels,
					},
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "lambda",
									Image:           imageName,
									Env:             append(instance.Spec.Env, envVarsForRevision...),
									Resources:       instance.Spec.Resources,
									ImagePullPolicy: corev1.PullIfNotPresent,
								},
							},
							ServiceAccountName: r.config.ImagePullAccountName,
						},
					},
				},
			},
		},
	}

	return service
}

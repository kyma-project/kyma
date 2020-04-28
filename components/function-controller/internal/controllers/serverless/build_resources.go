package serverless

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const (
	serviceBindingUsagesAnnotation = "servicebindingusages.servicecatalog.kyma-project.io/tracing-information"

	autoscalingKnativeMinScaleAnn = "autoscaling.knative.dev/minScale"
	autoscalingKnativeMaxScaleAnn = "autoscaling.knative.dev/maxScale"

	servingKnativeVisibilityLabel      = "serving.knative.dev/visibility"
	servingKnativeVisibilityLabelValue = "cluster-local"
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

func (r *FunctionReconciler) buildService(log logr.Logger, instance *serverlessv1alpha1.Function) servingv1.Service {
	imageName := r.buildExternalImageAddress(instance)
	serviceLabels := r.serviceLabels(instance)

	podAnnotations := r.servicePodAnnotations(instance)
	podLabels := r.servicePodLabels(log, instance)

	return servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName(),
			Namespace: instance.GetNamespace(),
			Labels:    serviceLabels,
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: podAnnotations,
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
	labels := make(map[string]string, len(instance.GetLabels())+3)
	for key, value := range instance.GetLabels() {
		labels[key] = value
	}

	labels[serverlessv1alpha1.FunctionNameLabel] = instance.Name
	labels[serverlessv1alpha1.FunctionManagedByLabel] = "function-controller"
	labels[serverlessv1alpha1.FunctionUUIDLabel] = string(instance.GetUID())

	return labels
}

func (r *FunctionReconciler) serviceLabels(instance *serverlessv1alpha1.Function) map[string]string {
	serviceLabels := r.functionLabels(instance)
	serviceLabels[servingKnativeVisibilityLabel] = servingKnativeVisibilityLabelValue
	return serviceLabels
}

func (r *FunctionReconciler) servicePodAnnotations(instance *serverlessv1alpha1.Function) map[string]string {
	annotations := map[string]string{
		autoscalingKnativeMinScaleAnn: "1",
		autoscalingKnativeMaxScaleAnn: "1",
	}
	if instance.Spec.MinReplicas != nil {
		annotations[autoscalingKnativeMinScaleAnn] = fmt.Sprintf("%d", *instance.Spec.MinReplicas)
	}
	if instance.Spec.MaxReplicas != nil {
		annotations[autoscalingKnativeMaxScaleAnn] = fmt.Sprintf("%d", *instance.Spec.MaxReplicas)
	}
	return annotations
}

func (r *FunctionReconciler) servicePodLabels(log logr.Logger, instance *serverlessv1alpha1.Function) map[string]string {
	functionLabels := r.functionLabels(instance)
	bindingLabels := r.retrieveBindingLabels(log, instance)
	podLabels := instance.Spec.PodLabels

	if podLabels == nil || len(podLabels) == 0 {
		for key, value := range bindingLabels {
			functionLabels[key] = value
		}
		return functionLabels
	}

	for key, value := range functionLabels {
		podLabels[key] = value
	}
	for key, value := range bindingLabels {
		podLabels[key] = value
	}

	return podLabels
}

func (r *FunctionReconciler) retrieveBindingLabels(log logr.Logger, instance *serverlessv1alpha1.Function) map[string]string {
	bindingLabels := map[string]string{}

	bindingAnnotation := instance.GetAnnotations()[serviceBindingUsagesAnnotation]
	if bindingAnnotation == "" {
		return bindingLabels
	}

	type binding map[string]map[string]map[string]string
	var bindings binding
	if err := json.Unmarshal([]byte(bindingAnnotation), &bindings); err != nil {
		log.Error(err, fmt.Sprintf("Cannot parse SeriveBindingUsage annotation %s", bindingAnnotation))
	}

	for _, service := range bindings {
		for key, value := range service["injectedLabels"] {
			bindingLabels[key] = value
		}
	}

	return bindingLabels
}

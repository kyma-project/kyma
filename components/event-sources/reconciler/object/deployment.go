package object

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewDeployment creates a Service object.
func NewDeployment(ns, name string, opts ...ObjectOption) *appsv1.Deployment {
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// WithName sets the container name of a Deployment
func WithName(name string) ObjectOption {
	return func(o metav1.Object) {
		d := o.(*appsv1.Deployment)
		firstDeploymentContainer(d).Name = name
	}
}

// WithMatchLabelsSelector sets the selector of a Deployment
func WithMatchLabelsSelector(key, value string) ObjectOption {
	return func(o metav1.Object) {
		d := o.(*appsv1.Deployment)
		d.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{key: value}}
	}
}

// WithImage sets the container image of a Service.
func WithImage(img string) ObjectOption {
	return func(o metav1.Object) {
		d := o.(*appsv1.Deployment)
		firstDeploymentContainer(d).Image = img
	}
}

// WithImage sets the container image of a Service.
func WithReplicas(replicas int) ObjectOption {
	return func(o metav1.Object) {
		d := o.(*appsv1.Deployment)
		replica32 := int32(replicas)
		d.Spec.Replicas = &replica32
	}
}

// WithPort sets the container port of a Service.
func WithPort(port int32, name string) ObjectOption {
	return func(o metav1.Object) {
		d := o.(*appsv1.Deployment)
		ports := &firstDeploymentContainer(d).Ports

		*ports = append(*ports, corev1.ContainerPort{
			ContainerPort: port,
			Name:          name,
		})
	}
}

// WithPodLabel sets a label on a Service's template
func WithPodLabel(key, val string) ObjectOption {
	return func(o metav1.Object) {
		d := o.(*appsv1.Deployment)

		tpl := d.Spec.Template
		if tpl.ObjectMeta.Labels == nil {
			tpl.ObjectMeta.Labels = make(map[string]string)
		}
		tpl.ObjectMeta.Labels[key] = val
		d.Spec.Template = tpl
	}
}

// WithEnvVar sets the value of a container env var.
func WithEnvVar(name, val string) ObjectOption {
	return func(o metav1.Object) {
		d := o.(*appsv1.Deployment)
		envvars := &firstDeploymentContainer(d).Env

		*envvars = append(*envvars, corev1.EnvVar{
			Name:  name,
			Value: val,
		})
	}
}

// WithProbe sets the HTTP readiness probe of a container.
func WithProbe(path string, port int) ObjectOption {
	return func(o metav1.Object) {
		d := o.(*appsv1.Deployment)

		firstDeploymentContainer(d).ReadinessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: path,
					Port: intstr.FromInt(port),
					// setting port explicitly is illegal in a Deployment
				},
			},
		}
	}
}

// firstDeploymentContainer returns the first Container definition of a Deployment. A
// new empty Container is injected if the Service does not contain any.
func firstDeploymentContainer(d *appsv1.Deployment) *corev1.Container {
	containers := &d.Spec.Template.Spec.Containers
	if *containers == nil {
		*containers = make([]corev1.Container, 1)
	}
	return &(*containers)[0]
}

// ApplyExistingDeploymentAttributes copies some important annotations from a given
// source Service to a destination Service.
func ApplyExistingDeploymentAttributes(src, dst *appsv1.Deployment) {
	// resourceVersion must be returned to the API server
	// unmodified for optimistic concurrency, as per Kubernetes API
	// conventions
	dst.ResourceVersion = src.ResourceVersion

	for _, ann := range deploymentAnnotations {
		if val, ok := src.Annotations[ann]; ok {
			metav1.SetMetaDataAnnotation(&dst.ObjectMeta, ann, val)
		}
	}

	// preserve status to avoid resetting conditions
	dst.Status = src.Status
}

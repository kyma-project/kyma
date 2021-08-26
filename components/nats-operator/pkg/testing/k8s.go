// Package testing provides support for automated testing of nats-operator-doctor.
package testing

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Option TODO ...
type Option func(runtime.Object)

// GetPod TODO ...
func GetPod(name, namespace string, opts ...Option) *v1.Pod {
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	for _, opt := range opts {
		opt(pod)
	}
	return pod
}

// GetPodIfNil TODO ...
func GetPodIfNil(pod *v1.Pod) *v1.Pod {
	if pod != nil {
		return pod
	}
	return &v1.Pod{}
}

// PodWithLabel TODO ...
func PodWithLabel(key, value string) Option {
	return func(object runtime.Object) {
		pod, ok := object.(*v1.Pod)
		if !ok {
			panic("object is not a kubernetes pod")
		}
		if pod.Labels == nil {
			pod.Labels = make(map[string]string, 1)
		}
		pod.Labels[key] = value
	}
}

// PodWithPhase TODO ...
func PodWithPhase(phase v1.PodPhase) Option {
	return func(object runtime.Object) {
		pod, ok := object.(*v1.Pod)
		if !ok {
			panic("object is not a kubernetes pod")
		}
		pod.Status.Phase = phase
	}
}

// PodsToRuntimeObjects TODO ...
func PodsToRuntimeObjects(pods []*v1.Pod) []runtime.Object {
	objects := make([]runtime.Object, 0, len(pods))
	for _, secret := range pods {
		objects = append(objects, secret)
	}
	return objects
}

// GetDeployment TODO ...
func GetDeployment(name, namespace string, opts ...Option) *appsv1.Deployment {
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	for _, opt := range opts {
		opt(deployment)
	}
	return deployment
}

// GetDeploymentIfNil TODO ...
func GetDeploymentIfNil(deployment *appsv1.Deployment) *appsv1.Deployment {
	if deployment != nil {
		return deployment
	}
	return &appsv1.Deployment{}
}

// DeploymentWithReplicas TODO ...
func DeploymentWithReplicas(replicas int32) Option {
	return func(object runtime.Object) {
		deployment, ok := object.(*appsv1.Deployment)
		if !ok {
			panic("object is not a kubernetes deployment")
		}
		deployment.Spec.Replicas = &replicas
	}
}

// GetSecret TODO ...
func GetSecret(name, namespace string, opts ...Option) *v1.Secret {
	secret := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	for _, opt := range opts {
		opt(secret)
	}
	return secret
}

// SecretWithLabel TODO ...
func SecretWithLabel(key, value string) Option {
	return func(object runtime.Object) {
		secret, ok := object.(*v1.Secret)
		if !ok {
			panic("object is not a kubernetes secret")
		}
		if secret.Labels == nil {
			secret.Labels = make(map[string]string, 1)
		}
		secret.Labels[key] = value
	}
}

// SecretsToRuntimeObjects TODO ...
func SecretsToRuntimeObjects(secrets []*v1.Secret) []runtime.Object {
	objects := make([]runtime.Object, 0, len(secrets))
	for _, secret := range secrets {
		objects = append(objects, secret)
	}
	return objects
}

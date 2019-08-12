package util

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// SubscriberName is the default subscriber name used in the tests
	SubscriberName = "test-core-event-bus-subscriber"
)

// NewSubscriberDeployment creates a new Kubernetes v1.Deployment instance with the default subscriber name.
func NewSubscriberDeployment(subscriberImage string) *appsv1.Deployment {
	return NewSubscriberDeploymentWithName(SubscriberName, subscriberImage)
}

// NewSubscriberDeploymentWithName creates a new Kubernetes v1.Deployment instance with the given subscriber name.
func NewSubscriberDeploymentWithName(subscriberName, subscriberImage string) *appsv1.Deployment {
	var replicas int32 = 1
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: subscriberName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": subscriberName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": subscriberName,
					},
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            subscriberName,
							Image:           subscriberImage,
							ImagePullPolicy: "IfNotPresent",
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 9000,
								},
							},
						},
					},
				},
			},
		},
	}
}

// NewSubscriberService creates a new Kubernetes v1.Service instance with the default subscriber name.
func NewSubscriberService() *apiv1.Service {
	return NewSubscriberServiceWithName(SubscriberName)
}

// NewSubscriberServiceWithName creates a new Kubernetes v1.Service instance with the given subscriber name.
func NewSubscriberServiceWithName(subscriberName string) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: subscriberName,
			Labels: map[string]string{
				"app": subscriberName,
			},
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"app": subscriberName,
			},
			Ports: []apiv1.ServicePort{
				{
					Name: "http",
					Port: 9000,
				},
			},
		},
	}
}

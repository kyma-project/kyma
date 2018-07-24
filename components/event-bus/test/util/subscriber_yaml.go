package util

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SubscriberName = "test-core-event-bus-subscriber"
)

func NewSubscriberDeployment(sbscrImg string) *appsv1.Deployment {
	var replicas int32 = 1
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: SubscriberName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": SubscriberName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": SubscriberName,
					},
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            SubscriberName,
							Image:           sbscrImg,
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

func NewSubscriberService() *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: SubscriberName,
			Labels: map[string]string{
				"app": SubscriberName,
			},
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"app": SubscriberName,
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

package helpers

import (
	"fmt"

	"github.com/avast/retry-go"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

// TODO(k15r): change image to a release image, not PR-XXX image
const subscriberImage = "eu.gcr.io/kyma-project/pr/event-bus-e2e-subscriber:PR-4893"

type SubscriberOption func(deployment *appsv1.Deployment)

func CreateSubscriber(k8s k8s.Interface, name, namespace string, subscriberOptions ...SubscriberOption) error {
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: name,

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

	for _, option := range subscriberOptions {
		option(deployment)
	}

	if _, err := k8s.AppsV1().Deployments(namespace).Create(deployment); err != nil {
		return err
	}

	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []apiv1.ServicePort{
				{
					Name: "http",
					Port: 9000,
				},
			},
		},
	}

	if _, err := k8s.CoreV1().Services(namespace).Create(service); err != nil {
		return err
	}
	return nil
}

func isPodReady(pod *apiv1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if !cs.Ready {
			return false
		}
	}
	return true
}

func WaitForSubscriber(k8s k8s.Interface, name, namespace string) error {
	return retry.Do(func() error {
		pods, err := k8s.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app=" + name})
		if err != nil {
			return err
		}
		for _, pod := range pods.Items {
			if !isPodReady(&pod) {
				return fmt.Errorf("subscriber pod is not ready: %v", pod)
			}
		}
		return nil
	})
}

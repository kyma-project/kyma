package installer

import (
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	corev1Type "k8s.io/client-go/kubernetes/typed/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	"fmt"
	"time"
)

const (
	podReadyTimeout = time.Second * 30
	UPSBrokerImage = "quay.io/kubernetes-service-catalog/user-broker:latest"
)

func CreatePod(k8sClient *corev1Type.CoreV1Client, namespace, name string) (*v1.Pod, error) {
	return k8sClient.Pods(namespace).Create(upsBrokerPod(name))
}

func DeletePod(k8sClient *corev1Type.CoreV1Client, namespace, name string) error {
	return k8sClient.Pods(namespace).Delete(name, nil)
}

func WaitForPodRunning(k8sClient *corev1Type.CoreV1Client, namespace, name string) error {
	return waiter.WaitAtMost(func() (bool, error) {
		pod, err := k8sClient.Pods(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case corev1.PodRunning:
			return true, nil
		case corev1.PodFailed, corev1.PodSucceeded:
			return false, fmt.Errorf("pod ran to completion")
		}

		return false, nil
	}, podReadyTimeout)
}

func upsBrokerPod(name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  name,
					Image: UPSBrokerImage,
					Args: []string{
						"--port",
						"8080",
						"-alsologtostderr",
					},
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 8080,
						},
					},
				},
			},
		},
	}
}
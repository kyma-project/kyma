package installer

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1Type "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	podReadyTimeout = time.Second * 60
)

type PodInstaller struct {
	name      string
	namespace string
}

func NewPod(name, namespace string) *PodInstaller {
	return &PodInstaller{
		name:      name,
		namespace: namespace,
	}
}

func (t *PodInstaller) Create(k8sClient *corev1Type.CoreV1Client, pod *corev1.Pod) error {
	_, err := k8sClient.Pods(t.namespace).Create(pod)
	return err
}

func (t *PodInstaller) Delete(k8sClient *corev1Type.CoreV1Client) error {
	return k8sClient.Pods(t.namespace).Delete(t.name, nil)
}

func (t *PodInstaller) Name() string {
	return t.name
}

func (t *PodInstaller) Namespace() string {
	return t.namespace
}

func (t *PodInstaller) WaitForPodRunning(k8sClient *corev1Type.CoreV1Client) error {
	return waiter.WaitAtMost(func() (bool, error) {
		pod, err := k8sClient.Pods(t.namespace).Get(t.name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case corev1.PodRunning:
			return true, nil
		case corev1.PodFailed, corev1.PodSucceeded:
			return false, fmt.Errorf("%v", pod.Status)
		}

		return false, nil
	}, podReadyTimeout)
}

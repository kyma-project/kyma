package installer

import (
	"time"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	corev1Type "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	endpointReadyTimeout = time.Second * 30
)

func CreateService(k8sClient *corev1Type.CoreV1Client, namespace, name string) (*v1.Service, error) {
	return k8sClient.Services(namespace).Create(upsBrokerService(name))
}

func DeleteService(k8sClient *corev1Type.CoreV1Client, namespace, name string) error {
	return k8sClient.Services(namespace).Delete(name, nil)
}

func WaitForEndpoint(k8sClient *corev1Type.CoreV1Client, namespace, name string) error {
	return waiter.WaitAtMost(func() (bool, error) {
		endpoint, err := k8sClient.Endpoints(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if len(endpoint.Subsets) == 0 || len(endpoint.Subsets[0].Addresses) == 0 {
			return false, nil
		}

		return true, nil
	}, endpointReadyTimeout)
}

func upsBrokerService(name string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}
}

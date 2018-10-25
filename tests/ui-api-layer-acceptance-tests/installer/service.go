package installer

import (
	"time"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	corev1Type "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	endpointReadyTimeout = time.Second * 60
)

type ServiceInstaller struct {
	name      string
	namespace string
}

func NewService(name, namespace string) *ServiceInstaller {
	return &ServiceInstaller{
		name:      name,
		namespace: namespace,
	}
}

func (t *ServiceInstaller) Create(k8sClient *corev1Type.CoreV1Client, service *corev1.Service) error {
	_, err := k8sClient.Services(t.namespace).Create(service)
	return err
}

func (t *ServiceInstaller) Delete(k8sClient *corev1Type.CoreV1Client) error {
	return k8sClient.Services(t.namespace).Delete(t.name, nil)
}

func (t *ServiceInstaller) Name() string {
	return t.name
}

func (t *ServiceInstaller) Namespace() string {
	return t.namespace
}

func (t *ServiceInstaller) WaitForEndpoint(k8sClient *corev1Type.CoreV1Client) error {
	return waiter.WaitAtMost(func() (bool, error) {
		endpoint, err := k8sClient.Endpoints(t.namespace).Get(t.name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if len(endpoint.Subsets) == 0 || len(endpoint.Subsets[0].Addresses) == 0 {
			return false, nil
		}

		return true, nil
	}, endpointReadyTimeout)
}

func UPSBrokerService(name string) *corev1.Service {
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

package gateway_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/gateway"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGatewayServiceProvider(t *testing.T) {
	// GIVEN
	fakeClientSet := fake.NewSimpleClientset(
		fixService("other1", "ysf-integration"),
		fixReService("prod", "ysf-integration", "ec-prod"),
		fixReService("stage", "ysf-integration", "ec-stage"),
		fixService("other1", "ysf-system"),
		fixReService("invalid-ec", "ysf-system", "ec-invalid"),
		fixPod(),
	)
	core := fakeClientSet.CoreV1()
	svc := gateway.NewProvider(core, "ysf-integration", time.Hour)
	stopCh := make(chan struct{})
	svc.WaitForCacheSync(stopCh)

	// WHEN
	items := svc.ListGatewayServices()

	// THEN
	assert.Len(t, items, 2)
	assert.Contains(t, items, gateway.ServiceData{
		Host: "prod.ysf-integration.svc.cluster.local:80",
		RemoteEnvironmentName: "ec-prod",
	})
	assert.Contains(t, items, gateway.ServiceData{
		Host: "stage.ysf-integration.svc.cluster.local:80",
		RemoteEnvironmentName: "ec-stage",
	})
}

func fixPod() *apiv1.Pod {
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dummy-pod",
			Namespace: "ysf-integration",
		},
		Spec: apiv1.PodSpec{},
	}
}

func fixService(name, namespace string) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{"app": "svc"},
			Ports: []apiv1.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
				},
			},
		},
	}
}

func fixReService(name, namespace, reName string) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"remoteEnvironment": reName,
			},
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{"app": "svc"},
			Ports: []apiv1.ServicePort{
				{
					Name:       "ext-api-port",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
				},
			},
		},
	}
}

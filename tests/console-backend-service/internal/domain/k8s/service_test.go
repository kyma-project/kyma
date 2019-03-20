// +build acceptance

package k8s

import (
	"testing"
	"time"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/dex"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/waiter"

	"k8s.io/apimachinery/pkg/util/intstr"

	_assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	serviceName = "test-service"
	namespace   = "ui-api-acceptance-service"
)

type ServiceEvent struct {
	Type    string
	Service service
}

type serviceQueryResponse struct {
	Service service `json:"service"`
}

type servicesQueryResponse struct {
	Services []service `json:"services"`
}

type service struct {
	Name              string        `json:"name"`
	ClusterIP         string        `json:"clusterIP"`
	CreationTimestamp int64         `json:"creationTimestamp"`
	Labels            labels        `json:"labels"`
	Ports             []ServicePort `json:"ports"`
	Status            ServiceStatus `json:"status"`
}

type ServiceStatus struct {
	LoadBalancer LoadBalancerStatus `json:"loadBalancer"`
}

type ServiceProtocol string

type ServicePort struct {
	Name            string          `json:"name"`
	ServiceProtocol ServiceProtocol `json:"serviceProtocol"`
	Port            int             `json:"port"`
	NodePort        int             `json:"nodePort"`
	TargetPort      int             `json:"targetPort"`
}

type LoadBalancerIngress struct {
	IP       string `json:"ip"`
	HostName string `json:"hostName"`
}

type LoadBalancerStatus struct {
	Ingress []LoadBalancerIngress `json:"ingress"`
}

func TestService(t *testing.T) {

	assert := _assert.New(t)
	dex.SkipTestIfSCIEnabled(t)

	grapqlClient, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	t.Log("Creating namespace...")
	_, err = k8sClient.Namespaces().Create(fixNamespace(namespace))
	require.NoError(t, err)
	defer func() {
		t.Log("Deleting namespace...")
		err = k8sClient.Namespaces().Delete(namespace, &metav1.DeleteOptions{})
		require.NoError(t, err)
	}()

	t.Log("Subscribing to serviceEvent...")
	subscription := grapqlClient.Subscribe(fixServicesSubscription())
	defer subscription.Close()

	t.Log("Creating service...")
	_, err = k8sClient.Services(namespace).Create(fixService(serviceName, namespace))
	require.NoError(t, err)

	t.Log("Retrieving service...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.Services(namespace).Get(serviceName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Checking subscription for created service...")
	expectedEvent := serviceEvent("ADD", service{Name: serviceName})
	assert.NoError(checkServiceEvent(expectedEvent, subscription))

	t.Log("Querying for service...")
	var serviceRes serviceQueryResponse
	err = grapqlClient.Do(fixServiceQuery(), &serviceRes)
	require.NoError(t, err)
	assert.Equal(serviceName, serviceRes.Service.Name)

	t.Log("Querying for services...")
	var servicesRes servicesQueryResponse
	err = grapqlClient.Do(fixServicesQuery(), &servicesRes)
	require.NoError(t, err)
	assert.Equal(servicesRes.Services[0].Name, serviceName)
}

func fixService(name, namespace string) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "a",
					Protocol:   v1.ProtocolTCP,
					Port:       1,
					TargetPort: intstr.FromInt(2),
				},
			},
		},
		Status: v1.ServiceStatus{},
	}
}

func fixServiceQuery() *graphql.Request {
	query := `query service($name: String!, $namespace: String!) {
  service(namespace:$namespace, name:$name) {
    name
    clusterIP
    creationTimestamp
    labels
    ports {
      name
      serviceProtocol
      port
      nodePort
      targetPort
    }
    status {
      loadBalancer {
        ingress {
          ip
          hostName
        }
      }
    }
  }
}`
	req := graphql.NewRequest(query)
	req.SetVar("name", serviceName)
	req.SetVar("namespace", namespace)

	return req
}

func fixServicesQuery() *graphql.Request {
	query := `query services($namespace: String!) {
  services(namespace: $namespace) {
    name
    clusterIP
    creationTimestamp
    labels
    ports {
      name
      serviceProtocol
      port
      nodePort
      targetPort
    }
    status{
      loadBalancer{
        ingress{
          ip
          hostName
        }
      }
    }
  }
}`
	req := graphql.NewRequest(query)
	req.SetVar("namespace", namespace)

	return req
}

func fixServicesSubscription() *graphql.Request {
	query := `subscription ($namespace: String!) {
serviceEvent(namespace: $namespace) {
    type
    service {
      name
      clusterIP
      creationTimestamp
      labels
      ports {
        name
        serviceProtocol
        port
        nodePort
        targetPort
      }
      status{
        loadBalancer{
          ingress{
            ip
            hostName
          }
        }
      }
    }
  }
}`
	req := graphql.NewRequest(query)
	req.SetVar("namespace", namespace)

	return req
}

func serviceEvent(eventType string, service service) ServiceEvent {
	return ServiceEvent{
		Type:    eventType,
		Service: service,
	}
}

func readServiceEvent(sub *graphql.Subscription) (ServiceEvent, error) {
	type Response struct {
		ServiceEvent ServiceEvent
	}
	var serviceEvent Response
	err := sub.Next(&serviceEvent, tester.DefaultSubscriptionTimeout)

	return serviceEvent.ServiceEvent, err
}

func checkServiceEvent(expected ServiceEvent, sub *graphql.Subscription) error {
	for {
		event, err := readServiceEvent(sub)
		if err != nil {
			return err
		}
		if expected.Type == event.Type && expected.Service.Name == event.Service.Name {
			return nil
		}
	}
}

// +build acceptance

package k8s

import (
	"testing"
	"time"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/retrier"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/waiter"

	_assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	serviceName = "test-service"
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

type updateServiceMutationResponse struct {
	UpdateService service `json:"updateService"`
}

type service struct {
	Name              string        `json:"name"`
	ClusterIP         string        `json:"clusterIP"`
	CreationTimestamp int64         `json:"creationTimestamp"`
	Labels            labels        `json:"labels"`
	Ports             []ServicePort `json:"ports"`
	Status            ServiceStatus `json:"status"`
	JSON              json          `json:"json"`
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

type deleteServiceMutationResponse struct {
	DeleteService service `json:"deleteService"`
}

func TestService(t *testing.T) {
	t.Skip("skipping unstable test")

	assert := _assert.New(t)
	grapqlClient, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	t.Log("Subscribing to serviceEvent...")
	subscription := grapqlClient.Subscribe(fixServicesSubscription())
	defer subscription.Close()

	t.Log("Creating service...")
	_, err = k8sClient.Services(testNamespace).Create(fixService(serviceName, testNamespace))
	require.NoError(t, err)

	t.Log("Retrieving service...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.Services(testNamespace).Get(serviceName, metav1.GetOptions{})
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

	t.Log("Updating...")
	var updateRes updateServiceMutationResponse
	err = retrier.Retry(func() error {
		err = grapqlClient.Do(fixServiceQuery(), &serviceRes)
		if err != nil {
			return err
		}
		serviceRes.Service.JSON["metadata"].(map[string]interface{})["labels"] = map[string]string{"foo": "bar"}
		update, err := stringifyJSON(serviceRes.Service.JSON)
		if err != nil {
			return err
		}
		err = grapqlClient.Do(fixUpdateServiceMutation(update), &updateRes)
		if err != nil {
			return err
		}
		return nil
	}, retrier.UpdateRetries)
	require.NoError(t, err)
	assert.Equal(serviceName, updateRes.UpdateService.Name)

	t.Log("Querying for services...")
	var servicesRes servicesQueryResponse
	err = grapqlClient.Do(fixServicesQuery(), &servicesRes)
	require.NoError(t, err)
	assert.Equal(servicesRes.Services[0].Name, serviceName)

	t.Log("Deleting service...")
	var deleteRes deleteServiceMutationResponse
	err = grapqlClient.Do(fixDeleteServiceMutation(), &deleteRes)
	require.NoError(t, err)
	assert.Equal(serviceName, deleteRes.DeleteService.Name)

	t.Log("Waiting for deletion...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.Services(testNamespace).Get(serviceName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Checking subscription for deleted service...")
	expectedEvent = serviceEvent("DELETE", service{Name: serviceName})
	assert.NoError(checkServiceEvent(expectedEvent, subscription))

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Get:    {fixServiceQuery()},
		auth.List:   {fixServicesQuery()},
		auth.Create: {fixUpdateServiceMutation("{\"\":\"\"}")},
		auth.Delete: {fixDeleteServiceMutation()},
	}
	AuthSuite.Run(t, ops)
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
	json
  }
}`
	req := graphql.NewRequest(query)
	req.SetVar("name", serviceName)
	req.SetVar("namespace", testNamespace)

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
	req.SetVar("namespace", testNamespace)

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
	req.SetVar("namespace", testNamespace)

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

func fixUpdateServiceMutation(service string) *graphql.Request {
	mutation := `mutation UpdateService($name: String!, $namespace: String!, $service: JSON!) {
  updateService(name: $name, namespace: $namespace, service: $service) {
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
	req := graphql.NewRequest(mutation)
	req.SetVar("name", serviceName)
	req.SetVar("namespace", testNamespace)
	req.SetVar("service", service)

	return req
}

func fixDeleteServiceMutation() *graphql.Request {
	mutation := `mutation DeleteService($name: String!, $namespace: String!) {
  deleteService(name: $name, namespace: $namespace){
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
	req := graphql.NewRequest(mutation)
	req.SetVar("name", serviceName)
	req.SetVar("namespace", testNamespace)
	return req
}

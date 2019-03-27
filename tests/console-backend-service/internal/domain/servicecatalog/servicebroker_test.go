// +build acceptance

package servicecatalog

import (
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/wait"
	"github.com/pkg/errors"
	"log"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceBroker struct {
	Name              string
	Namespace         string
	CreationTimestamp int
	Url               string
	Labels            map[string]interface{}
	Status            ServiceBrokerStatus
}

type ServiceBrokerStatus struct {
	Ready   bool
	Reason  string
	Message string
}

type serviceBrokersQueryResponse struct {
	ServiceBrokers []ServiceBroker
}

type serviceBrokerQueryResponse struct {
	ServiceBroker ServiceBroker
}

func TestServiceBrokerQueries(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	expectedResource := broker()

	svcatCli, _, err := client.NewServiceCatalogClientWithConfig()
	require.NoError(t, err)

	testBroker := newTestServiceBroker(expectedResource.Name, expectedResource.Namespace, svcatCli)
	err = testBroker.Create()
	require.NoError(t, err)

	defer func() {
		err := testBroker.Delete()
		if err != nil {
			log.Printf(errors.Wrapf(err, "while deleting test ServiceBroker").Error())
		}
	}()

	err = wait.ForServiceBroker(expectedResource.Name, expectedResource.Namespace, svcatCli)
	require.NoError(t, err)

	resourceDetailsQuery := `
		name
		namespace
		creationTimestamp
		url
		labels
		status {
			ready
			reason
			message
		}
	`

	t.Run("MultipleResources", func(t *testing.T) {
		query := fmt.Sprintf(`
			query ($namespace: String!) {
				serviceBrokers(namespace: $namespace) {
					%s
				}
			}	
		`, resourceDetailsQuery)
		req := graphql.NewRequest(query)
		req.SetVar("namespace", expectedResource.Namespace)

		var res serviceBrokersQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		assertBrokerExistsAndEqual(t, res.ServiceBrokers, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		query := fmt.Sprintf(`
			query ($name: String!, $namespace: String!) {
				serviceBroker(name: $name, namespace: $namespace) {
					%s
				}
			}
		`, resourceDetailsQuery)
		req := graphql.NewRequest(query)
		req.SetVar("name", expectedResource.Name)
		req.SetVar("namespace", expectedResource.Namespace)

		var res serviceBrokerQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		checkBroker(t, expectedResource, res.ServiceBroker)
	})
}

func checkBroker(t *testing.T, expected, actual ServiceBroker) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// Url
	assert.Contains(t, actual.Url, expected.Url)

	// Namespace
	assert.Equal(t, expected.Namespace, actual.Namespace)
}

func assertBrokerExistsAndEqual(t *testing.T, arr []ServiceBroker, expectedElement ServiceBroker) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkBroker(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

type testServiceBroker struct {
	name      string
	namespace string
	svcatCli  *clientset.Clientset
}

func newTestServiceBroker(name string, namespace string, svcatCli *clientset.Clientset) *testServiceBroker {
	return &testServiceBroker{name: name, namespace: namespace, svcatCli: svcatCli}
}

func (t *testServiceBroker) Create() error {
	broker := &v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.name,
			Namespace: t.namespace,
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: CommonBrokerURL,
			},
		},
	}

	_, err := t.svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.namespace).Create(broker)
	return err
}

func (t *testServiceBroker) Delete() error {
	return t.svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.namespace).Delete(t.name, &metav1.DeleteOptions{GracePeriodSeconds: &brokerDeletionGracefulPeriod})
}

func broker() ServiceBroker {
	return ServiceBroker{
		Name:      BrokerReleaseName,
		Namespace: TestNamespace,
		Url: CommonBrokerURL,
	}
}

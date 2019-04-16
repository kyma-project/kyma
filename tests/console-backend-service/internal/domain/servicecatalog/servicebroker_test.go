// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/fixture"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		req := fixServiceBrokersRequest(resourceDetailsQuery, expectedResource)

		var res serviceBrokersQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		assertBrokerExistsAndEqual(t, res.ServiceBrokers, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		req := fixServiceBrokerRequest(resourceDetailsQuery, expectedResource)

		var res serviceBrokerQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		checkBroker(t, expectedResource, res.ServiceBroker)
	})

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Get:  {fixServiceBrokerRequest(resourceDetailsQuery, expectedResource)},
		auth.List: {fixServiceBrokersRequest(resourceDetailsQuery, expectedResource)},
	}
	AuthSuite.Run(t, ops)
}

func checkBroker(t *testing.T, expected, actual ServiceBroker) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

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

func broker() ServiceBroker {
	return ServiceBroker{
		Name:      fmt.Sprintf("ns-%s", fixture.TestingBrokerName),
		Namespace: TestNamespace,
	}
}

func fixServiceBrokerRequest(resourceDetailsQuery string, expectedResource ServiceBroker) *graphql.Request {
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

	return req
}

func fixServiceBrokersRequest(resourceDetailsQuery string, expectedResource ServiceBroker) *graphql.Request {
	query := fmt.Sprintf(`
			query ($namespace: String!) {
				serviceBrokers(namespace: $namespace) {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("namespace", expectedResource.Namespace)

	return req
}

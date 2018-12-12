// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ServiceBroker struct {
	Name              string
	Environment       string
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
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}

	c, err := graphql.New()
	require.NoError(t, err)

	expectedResource := broker()
	resourceDetailsQuery := `
		name
		environment
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
			query ($environment: String!) {
				serviceBrokers(environment: $environment) {
					%s
				}
			}	
		`, resourceDetailsQuery)
		req := graphql.NewRequest(query)
		req.SetVar("environment", expectedResource.Environment)

		var res serviceBrokersQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		assertBrokerExistsAndEqual(t, res.ServiceBrokers, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		query := fmt.Sprintf(`
			query ($name: String!, $environment: String!) {
				serviceBroker(name: $name, environment: $environment) {
					%s
				}
			}
		`, resourceDetailsQuery)
		req := graphql.NewRequest(query)
		req.SetVar("name", expectedResource.Name)
		req.SetVar("environment", expectedResource.Environment)

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
	assert.Contains(t, actual.Url, expected.Name)

	// Environment
	assert.Equal(t, expected.Environment, actual.Environment)

	// Status
	assert.Equal(t, expected.Status.Ready, actual.Status.Ready)
	assert.NotEmpty(t, actual.Status.Message)
	assert.NotEmpty(t, actual.Status.Reason)
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
		Name:        tester.BrokerReleaseName,
		Environment: tester.DefaultNamespace,
		Status: ServiceBrokerStatus{
			Ready: true,
		},
	}
}

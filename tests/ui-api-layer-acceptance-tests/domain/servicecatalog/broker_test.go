// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ServiceBroker struct {
	Name              string
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

type brokersQueryResponse struct {
	ServiceBrokers []ServiceBroker
}

type brokerQueryResponse struct {
	ServiceBroker ServiceBroker
}

func TestBrokerQueries(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}

	c, err := graphql.New()
	require.NoError(t, err)

	expectedResource := broker()
	resourceDetailsQuery := `
		name
		creationTimestamp
    	status {
			ready
    		reason
    		message
		}
    	url
    	labels
	`

	t.Run("MultipleResources", func(t *testing.T) {
		query := fmt.Sprintf(`
			query {
				serviceBrokers {
					%s
				}
			}	
		`, resourceDetailsQuery)

		var res brokersQueryResponse
		err = c.DoQuery(query, &res)

		require.NoError(t, err)
		assertBrokerExistsAndEqual(t, res.ServiceBrokers, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		query := fmt.Sprintf(`
			query($name: String!) {
				serviceBroker(name: $name) {
					%s
				}
			}
		`, resourceDetailsQuery)
		req := graphql.NewRequest(query)
		req.SetVar("name", expectedResource.Name)

		var res brokerQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		checkBroker(t, expectedResource, res.ServiceBroker)
	})
}

func checkBroker(t *testing.T, expected, actual ServiceBroker) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Contains(t, actual.Url, expected.Name)

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
		Name: "ups-broker",
		Status: ServiceBrokerStatus{
			Ready: true,
		},
	}
}

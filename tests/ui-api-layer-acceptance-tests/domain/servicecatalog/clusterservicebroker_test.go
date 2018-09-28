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

type ClusterServiceBroker struct {
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
	ClusterServiceBrokers []ClusterServiceBroker
}

type brokerQueryResponse struct {
	ClusterServiceBroker ClusterServiceBroker
}

func TestClusterServiceBrokerQueries(t *testing.T) {
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
				clusterServiceBrokers {
					%s
				}
			}	
		`, resourceDetailsQuery)

		var res brokersQueryResponse
		err = c.DoQuery(query, &res)

		require.NoError(t, err)
		assertBrokerExistsAndEqual(t, res.ClusterServiceBrokers, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		query := fmt.Sprintf(`
			query($name: String!) {
				clusterServiceBroker(name: $name) {
					%s
				}
			}
		`, resourceDetailsQuery)
		req := graphql.NewRequest(query)
		req.SetVar("name", expectedResource.Name)

		var res brokerQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		checkBroker(t, expectedResource, res.ClusterServiceBroker)
	})
}

func checkBroker(t *testing.T, expected, actual ClusterServiceBroker) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Contains(t, actual.Url, expected.Name)

	// Status
	assert.Equal(t, expected.Status.Ready, actual.Status.Ready)
	assert.NotEmpty(t, actual.Status.Message)
	assert.NotEmpty(t, actual.Status.Reason)
}

func assertBrokerExistsAndEqual(t *testing.T, arr []ClusterServiceBroker, expectedElement ClusterServiceBroker) {
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

func broker() ClusterServiceBroker {
	return ClusterServiceBroker{
		Name: "ups-broker",
		Status: ServiceBrokerStatus{
			Ready: true,
		},
	}
}

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

type ClusterServiceBroker struct {
	Name              string
	CreationTimestamp int
	Url               string
	Labels            map[string]interface{}
	Status            ClusterServiceBrokerStatus
}

type ClusterServiceBrokerStatus struct {
	Ready   bool
	Reason  string
	Message string
}

type clusterServiceBrokersQueryResponse struct {
	ClusterServiceBrokers []ClusterServiceBroker
}

type clusterServiceBrokerQueryResponse struct {
	ClusterServiceBroker ClusterServiceBroker
}

func TestClusterServiceBrokerQueries(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}

	c, err := graphql.New()
	require.NoError(t, err)

	expectedResource := clusterBroker()
	resourceDetailsQuery := `
		name
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
			query {
				clusterServiceBrokers {
					%s
				}
			}	
		`, resourceDetailsQuery)

		var res clusterServiceBrokersQueryResponse
		err = c.DoQuery(query, &res)

		require.NoError(t, err)
		assertClusterBrokerExistsAndEqual(t, res.ClusterServiceBrokers, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		query := fmt.Sprintf(`
			query ($name: String!) {
				clusterServiceBroker(name: $name) {
					%s
				}
			}
		`, resourceDetailsQuery)
		req := graphql.NewRequest(query)
		req.SetVar("name", expectedResource.Name)

		var res clusterServiceBrokerQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		checkClusterBroker(t, expectedResource, res.ClusterServiceBroker)
	})
}

func checkClusterBroker(t *testing.T, expected, actual ClusterServiceBroker) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// Url
	assert.Contains(t, actual.Url, expected.Name)

	// Status
	assert.Equal(t, expected.Status.Ready, actual.Status.Ready)
	assert.NotEmpty(t, actual.Status.Message)
	assert.NotEmpty(t, actual.Status.Reason)
}

func assertClusterBrokerExistsAndEqual(t *testing.T, arr []ClusterServiceBroker, expectedElement ClusterServiceBroker) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkClusterBroker(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

func clusterBroker() ClusterServiceBroker {
	return ClusterServiceBroker{
		Name: tester.ClusterBrokerReleaseName,
		Status: ClusterServiceBrokerStatus{
			Ready: true,
		},
	}
}

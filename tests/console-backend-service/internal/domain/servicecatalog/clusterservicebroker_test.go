// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/fixture"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"

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
		req := fixClusterServiceBrokersRequest(resourceDetailsQuery)

		var res clusterServiceBrokersQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		assertClusterBrokerExistsAndEqual(t, res.ClusterServiceBrokers, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		req := fixClusterServiceBrokerRequest(resourceDetailsQuery, expectedResource.Name)

		var res clusterServiceBrokerQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		checkClusterBroker(t, expectedResource, res.ClusterServiceBroker)
	})

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Get:  {fixClusterServiceBrokerRequest(resourceDetailsQuery, "test")},
		auth.List: {fixClusterServiceBrokersRequest(resourceDetailsQuery)},
	}
	AuthSuite.Run(t, ops)
}

func checkClusterBroker(t *testing.T, expected, actual ClusterServiceBroker) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// Status
	assert.Equal(t, expected.Status.Ready, actual.Status.Ready)
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
		Name: fixture.TestingBrokerName,
		Status: ClusterServiceBrokerStatus{
			Ready: true,
		},
	}
}

func fixClusterServiceBrokerRequest(resourceDetailsQuery, name string) *graphql.Request {
	query := fmt.Sprintf(`
			query ($name: String!) {
				clusterServiceBroker(name: $name) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", name)

	return req
}

func fixClusterServiceBrokersRequest(resourceDetailsQuery string) *graphql.Request {
	query := fmt.Sprintf(`
			query {
				clusterServiceBrokers {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)

	return req
}

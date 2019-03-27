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

	svcatCli, _, err := client.NewServiceCatalogClientWithConfig()
	require.NoError(t, err)

	testBroker := newTestClusterServiceBroker(expectedResource.Name, svcatCli)
	err = testBroker.Create()
	require.NoError(t, err)

	defer func() {
		err := testBroker.Delete()
		if err != nil {
			log.Printf(errors.Wrapf(err, "while deleting test ServiceBroker").Error())
		}
	}()

	err = wait.ForClusterServiceBroker(expectedResource.Name, svcatCli)
	require.NoError(t, err)

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
	assert.Equal(t, expected.Url, actual.Url)
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
		Name: ClusterBrokerReleaseName,
		Status: ClusterServiceBrokerStatus{
			Ready: true,
		},
		Url: CommonBrokerURL,
	}
}

type testClusterServiceBroker struct {
	name     string
	svcatCli *clientset.Clientset
}

func newTestClusterServiceBroker(name string, svcatCli *clientset.Clientset) *testClusterServiceBroker {
	return &testClusterServiceBroker{name: name, svcatCli: svcatCli}
}

func (t *testClusterServiceBroker) Create() error {
	broker := &v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.name,
		},
		Spec: v1beta1.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: CommonBrokerURL,
			},
		},
	}

	_, err := t.svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Create(broker)
	return err
}

func (t *testClusterServiceBroker) Delete() error {
	return t.svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Delete(t.name, &metav1.DeleteOptions{GracePeriodSeconds: &brokerDeletionGracefulPeriod})
}

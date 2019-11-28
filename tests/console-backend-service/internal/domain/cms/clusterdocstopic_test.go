// +build acceptance

package cms

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/fixture"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/wait"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/mockice"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	clusterDocsTopicName1 = "example-cluster-docs-topic-1"
	clusterDocsTopicName2 = "example-cluster-docs-topic-2"
	clusterDocsTopicName3 = "example-cluster-docs-topic-3"
)

type ClusterDocsTopicEvent struct {
	Type             string
	ClusterDocsTopic shared.ClusterDocsTopic
}

type clusterDocsTopicsQueryResponse struct {
	ClusterDocsTopics []shared.ClusterDocsTopic
}

func TestClusterDocsTopicsQueries(t *testing.T) {
	t.Skip("skipping unstable test")
	c, err := graphql.New()
	require.NoError(t, err)

	cmsCli, _, err := client.NewDynamicClientWithConfig()
	require.NoError(t, err)

	t.Log("Setup test service")
	host, err := mockice.Start(cmsCli, MockiceNamespace, MockiceSvcName)
	require.NoError(t, err)
	defer mockice.Stop(cmsCli, MockiceNamespace, MockiceSvcName)

	subscription := subscribeClusterDocsTopic(c, clusterDocsTopicEventDetailsFields())
	defer subscription.Close()

	clusterDocsTopicClient := resource.NewClusterDocsTopic(cmsCli, t.Logf)

	createClusterDocsTopic(t, clusterDocsTopicClient, clusterDocsTopicName1, "1", host)
	fixedClusterDocsTopic := fixture.ClusterDocsTopic(clusterDocsTopicName1)

	t.Log(fmt.Sprintf("Check subscription event of clusterDocsTopic %s created", clusterDocsTopicName1))
	expectedEvent := clusterDocsTopicEvent("ADD", fixedClusterDocsTopic)
	event, err := readClusterDocsTopicEvent(subscription)
	assert.NoError(t, err)
	checkClusterDocsTopicEvent(t, expectedEvent, event)

	createClusterDocsTopic(t, clusterDocsTopicClient, clusterDocsTopicName3, "3", host)
	createClusterDocsTopic(t, clusterDocsTopicClient, clusterDocsTopicName2, "2", host)

	waitForClusterDocsTopic(t, clusterDocsTopicClient, clusterDocsTopicName1)
	waitForClusterDocsTopic(t, clusterDocsTopicClient, clusterDocsTopicName3)
	waitForClusterDocsTopic(t, clusterDocsTopicClient, clusterDocsTopicName2)

	t.Log("Query Multiple Resources")
	multipleRes, err := queryMultipleClusterDocsTopics(c, clusterDocsTopicDetailsFields())
	assert.NoError(t, err)
	assert.Equal(t, 3, len(multipleRes.ClusterDocsTopics))
	assertClusterDocsTopicExistsAndEqual(t, fixedClusterDocsTopic, multipleRes.ClusterDocsTopics)

	deleteClusterDocsTopics(t, clusterDocsTopicClient)

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.List: {fixClusterDocsTopicsQuery(clusterDocsTopicDetailsFields())},
	}
	AuthSuite.Run(t, ops)
}

func createClusterDocsTopic(t *testing.T, client *resource.ClusterDocsTopic, name, order, host string) {
	t.Log(fmt.Sprintf("Create clusterDocsTopic %s", name))
	err := client.Create(fixClusterDocsTopicMeta(name, order), fixCommonClusterDocsTopicSpec(host))
	require.NoError(t, err)
}

func waitForClusterDocsTopic(t *testing.T, client *resource.ClusterDocsTopic, name string) {
	t.Log(fmt.Sprintf("Wait for clusterDocsTopic %s Ready", name))
	err := wait.ForClusterDocsTopicReady(name, client.Get)
	require.NoError(t, err)
}

func deleteClusterDocsTopics(t *testing.T, client *resource.ClusterDocsTopic) {
	t.Log("Deleting Cluster Docs Topics")
	dtNames := []string{
		clusterDocsTopicName1,
		clusterDocsTopicName2,
		clusterDocsTopicName3,
	}
	for _, name := range dtNames {
		err := client.Delete(name)
		assert.NoError(t, err)
	}
}

func fixClusterDocsTopicsQuery(resourceDetailsQuery string) *graphql.Request {
	query := fmt.Sprintf(`
			query ($viewContext: String, $groupName: String) {
				clusterDocsTopics (viewContext: $viewContext, groupName: $groupName) {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("viewContext", fixture.DocsTopicViewContext)
	req.SetVar("groupName", fixture.DocsTopicGroupName)

	return req
}

func queryMultipleClusterDocsTopics(c *graphql.Client, resourceDetailsQuery string) (clusterDocsTopicsQueryResponse, error) {
	req := fixClusterDocsTopicsQuery(resourceDetailsQuery)

	var res clusterDocsTopicsQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func assertClusterDocsTopicExistsAndEqual(t *testing.T, expectedElement shared.ClusterDocsTopic, arr []shared.ClusterDocsTopic) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkClusterDocsTopic(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "ClusterDocsTopic does not exist")
}

func assertClusterAssetsExistsAndEqual(t *testing.T, expectedElement shared.ClusterAsset, arr []shared.ClusterAsset) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if strings.HasPrefix(v.Name, expectedElement.Name) {
				checkClusterAsset(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "ClusterAsset does not exist")
}

func checkClusterDocsTopic(t *testing.T, expected, actual shared.ClusterDocsTopic) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// GroupName
	assert.Equal(t, expected.GroupName, actual.GroupName)

	// DisplayName
	assert.Equal(t, expected.DisplayName, actual.DisplayName)

	// Description
	assert.Equal(t, expected.Description, actual.Description)

	// Assets
	assertClusterAssetsExistsAndEqual(t, fixture.ClusterAsset(SourceType), actual.Assets)
}

func checkClusterAsset(t *testing.T, expected, actual shared.ClusterAsset) {
	// Type
	assert.Equal(t, expected.Type, actual.Type)

	// Length of Files
	assert.Equal(t, 1, len(actual.Files))
}

func subscribeClusterDocsTopic(c *graphql.Client, resourceDetailsQuery string) *graphql.Subscription {
	query := fmt.Sprintf(`
		subscription {
			clusterDocsTopicEvent {
				%s
			}
		}
	`, resourceDetailsQuery)
	req := graphql.NewRequest(query)

	return c.Subscribe(req)
}

func clusterDocsTopicDetailsFields() string {
	return `
		name
    	groupName
    	assets {
			name
			metadata
			parameters
			type
			files {
				url
				metadata
			}
		}
    	displayName
    	description
	`
}

func clusterDocsTopicEventDetailsFields() string {
	return fmt.Sprintf(`
        type
        clusterDocsTopic {
			%s
        }
    `, clusterDocsTopicDetailsFields())
}

func clusterDocsTopicEvent(eventType string, clusterDocsTopic shared.ClusterDocsTopic) ClusterDocsTopicEvent {
	return ClusterDocsTopicEvent{
		Type:             eventType,
		ClusterDocsTopic: clusterDocsTopic,
	}
}

func readClusterDocsTopicEvent(sub *graphql.Subscription) (ClusterDocsTopicEvent, error) {
	type Response struct {
		ClusterDocsTopicEvent ClusterDocsTopicEvent
	}

	var clusterDocsTopicEvent Response
	err := sub.Next(&clusterDocsTopicEvent, tester.DefaultSubscriptionTimeout)

	return clusterDocsTopicEvent.ClusterDocsTopicEvent, err
}

func checkClusterDocsTopicEvent(t *testing.T, expected, actual ClusterDocsTopicEvent) {
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.ClusterDocsTopic.Name, actual.ClusterDocsTopic.Name)
}

func fixClusterDocsTopicMeta(name, order string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: name,
		Labels: map[string]string{
			ViewContextLabel: fixture.DocsTopicViewContext,
			GroupNameLabel:   fixture.DocsTopicGroupName,
			OrderLabel:       order,
		},
	}
}

func fixCommonClusterDocsTopicSpec(host string) v1alpha1.CommonDocsTopicSpec {
	return v1alpha1.CommonDocsTopicSpec{
		DisplayName: fixture.DocsTopicDisplayName,
		Description: fixture.DocsTopicDescription,
		Sources: []v1alpha1.Source{
			{
				Type:       SourceType,
				Name:       SourceType,
				Parameters: &runtime.RawExtension{Raw: []byte(`{"json":"true","complex":{"data":"true"}}`)},
				Mode:       v1alpha1.DocsTopicSingle,
				URL:        mockice.ResourceURL(host),
			},
		},
	}
}

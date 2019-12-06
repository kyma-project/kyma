// +build acceptance

package rafter

import (
	"fmt"
	"strings"
	"testing"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/fixture"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/wait"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/mockice"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/resource"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	clusterAssetGroupName1 = "example-cluster-asset-group-1"
	clusterAssetGroupName2 = "example-cluster-asset-group-2"
	clusterAssetGroupName3 = "example-cluster-asset-group-3"
)

type clusterAssetGroupEvent struct {
	Type              string
	ClusterAssetGroup shared.ClusterAssetGroup
}

type clusterAssetGroupsQueryResponse struct {
	ClusterAssetGroups []shared.ClusterAssetGroup
}

func TestClusterAssetGroupsQueries(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	rafterCli, _, err := client.NewDynamicClientWithConfig()
	require.NoError(t, err)

	t.Log("Setup test service")
	host, err := mockice.Start(rafterCli, MockiceNamespace, MockiceSvcName)
	require.NoError(t, err)
	defer mockice.Stop(rafterCli, MockiceNamespace, MockiceSvcName)

	subscription := subscribeClusterAssetGroup(c, clusterAssetGroupEventDetailsFields())
	defer subscription.Close()

	clusterAssetGroupClient := resource.NewClusterAssetGroup(rafterCli, t.Logf)

	createClusterAssetGroup(t, clusterAssetGroupClient, clusterAssetGroupName1, "1", host)
	fixedClusterAssetGroup := fixture.ClusterAssetGroup(clusterAssetGroupName1)

	t.Log(fmt.Sprintf("Check subscription event of ClusterAssetGroup %s created", clusterAssetGroupName1))
	expectedEvent := newClusterAssetGroupEvent("ADD", fixedClusterAssetGroup)
	event, err := readClusterAssetGroupEvent(subscription)
	assert.NoError(t, err)
	checkClusterAssetGroupEvent(t, expectedEvent, event)

	createClusterAssetGroup(t, clusterAssetGroupClient, clusterAssetGroupName3, "3", host)
	createClusterAssetGroup(t, clusterAssetGroupClient, clusterAssetGroupName2, "2", host)

	waitForClusterAssetGroup(t, clusterAssetGroupClient, clusterAssetGroupName1)
	waitForClusterAssetGroup(t, clusterAssetGroupClient, clusterAssetGroupName3)
	waitForClusterAssetGroup(t, clusterAssetGroupClient, clusterAssetGroupName2)

	t.Log("Query Multiple Resources")
	multipleRes, err := queryMultipleClusterAssetGroups(c, clusterAssetGroupDetailsFields())
	assert.NoError(t, err)
	assert.Equal(t, 3, len(multipleRes.ClusterAssetGroups))
	assertClusterAssetGroupExistsAndEqual(t, fixedClusterAssetGroup, multipleRes.ClusterAssetGroups)

	deleteClusterAssetGroups(t, clusterAssetGroupClient)

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.List: {fixClusterAssetGroupsQuery(clusterAssetGroupDetailsFields())},
	}
	AuthSuite.Run(t, ops)
}

func createClusterAssetGroup(t *testing.T, client *resource.ClusterAssetGroup, name, order, host string) {
	t.Log(fmt.Sprintf("Create ClusterAssetGroup %s", name))
	err := client.Create(fixClusterAssetGroupMeta(name, order), fixCommonClusterAssetGroupSpec(host))
	require.NoError(t, err)
}

func waitForClusterAssetGroup(t *testing.T, client *resource.ClusterAssetGroup, name string) {
	t.Log(fmt.Sprintf("Wait for ClusterAssetGroup %s Ready", name))
	err := wait.ForClusterAssetGroupReady(name, client.Get)
	require.NoError(t, err)
}

func deleteClusterAssetGroups(t *testing.T, client *resource.ClusterAssetGroup) {
	t.Log("Deleting ClusterAssetGroups")
	dtNames := []string{
		clusterAssetGroupName1,
		clusterAssetGroupName2,
		clusterAssetGroupName3,
	}
	for _, name := range dtNames {
		err := client.Delete(name)
		assert.NoError(t, err)
	}
}

func fixClusterAssetGroupsQuery(resourceDetailsQuery string) *graphql.Request {
	query := fmt.Sprintf(`
			query ($viewContext: String, $groupName: String) {
				clusterAssetGroups (viewContext: $viewContext, groupName: $groupName) {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("viewContext", fixture.AssetGroupViewContext)
	req.SetVar("groupName", fixture.AssetGroupGroupName)

	return req
}

func queryMultipleClusterAssetGroups(c *graphql.Client, resourceDetailsQuery string) (clusterAssetGroupsQueryResponse, error) {
	req := fixClusterAssetGroupsQuery(resourceDetailsQuery)

	var res clusterAssetGroupsQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func assertClusterAssetGroupExistsAndEqual(t *testing.T, expectedElement shared.ClusterAssetGroup, arr []shared.ClusterAssetGroup) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkClusterAssetGroup(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "ClusterAssetGroup does not exist")
}

func assertClusterAssetsExistsAndEqual(t *testing.T, expectedElement shared.ClusterAsset, arr []shared.ClusterAsset) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if strings.HasPrefix(string(v.Name), string(expectedElement.Name)) {
				checkClusterAsset(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "ClusterAsset does not exist")
}

func checkClusterAssetGroup(t *testing.T, expected, actual shared.ClusterAssetGroup) {
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

func subscribeClusterAssetGroup(c *graphql.Client, resourceDetailsQuery string) *graphql.Subscription {
	query := fmt.Sprintf(`
		subscription {
			clusterAssetGroupEvent {
				%s
			}
		}
	`, resourceDetailsQuery)
	req := graphql.NewRequest(query)

	return c.Subscribe(req)
}

func clusterAssetGroupDetailsFields() string {
	return `
		name
    	groupName
    	assets {
			name
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

func clusterAssetGroupEventDetailsFields() string {
	return fmt.Sprintf(`
        type
        clusterAssetGroup {
			%s
        }
    `, clusterAssetGroupDetailsFields())
}

func newClusterAssetGroupEvent(eventType string, clusterAssetGroup shared.ClusterAssetGroup) clusterAssetGroupEvent {
	return clusterAssetGroupEvent{
		Type:              eventType,
		ClusterAssetGroup: clusterAssetGroup,
	}
}

func readClusterAssetGroupEvent(sub *graphql.Subscription) (clusterAssetGroupEvent, error) {
	type Response struct {
		ClusterAssetGroupEvent clusterAssetGroupEvent
	}

	var clusterAssetGroupEvent Response
	err := sub.Next(&clusterAssetGroupEvent, tester.DefaultSubscriptionTimeout)

	return clusterAssetGroupEvent.ClusterAssetGroupEvent, err
}

func checkClusterAssetGroupEvent(t *testing.T, expected, actual clusterAssetGroupEvent) {
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.ClusterAssetGroup.Name, actual.ClusterAssetGroup.Name)
}

func fixClusterAssetGroupMeta(name, order string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: name,
		Labels: map[string]string{
			ViewContextLabel: fixture.AssetGroupViewContext,
			GroupNameLabel:   fixture.AssetGroupGroupName,
			OrderLabel:       order,
		},
	}
}

func fixCommonClusterAssetGroupSpec(host string) v1beta1.CommonAssetGroupSpec {
	return v1beta1.CommonAssetGroupSpec{
		DisplayName: fixture.AssetGroupDisplayName,
		Description: fixture.AssetGroupDescription,
		Sources: []v1beta1.Source{
			{
				Type:       SourceType,
				Name:       SourceName,
				Parameters: &runtime.RawExtension{Raw: []byte(`{"json":"true","complex":{"data":"true"}}`)},
				Mode:       v1beta1.AssetGroupSingle,
				URL:        mockice.ResourceURL(host),
			},
		},
	}
}

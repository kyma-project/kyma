// +build acceptance

package remoteenvironment

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"

	clientset "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/client"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	reReadyTimeout = time.Second * 45
)

type remoteEnvironment struct {
	Name                  string
	Description           string
	Labels                map[string]string
	Services              []remoteEnvironmentService
	EnabledInEnvironments []string
	Status                string
}

type remoteEnvironmentService struct {
	Id                  string
	DisplayName         string
	LongDescription     string
	ProviderDisplayName string
	Tags                string
	Entries             remoteEnvironmentEntry
}

type remoteEnvironmentEntry struct {
	Type        string
	GatewayUrl  string
	AccessLabel string
}

type reDeleteMutation struct {
	Name string `json:"name"`
}

type reCreateMutationResponse struct {
	ReCreateMutation remoteEnvironment `json:"createRemoteEnvironment"`
}

type reUpdateMutationResponse struct {
	ReUpdateMutation remoteEnvironment `json:"updateRemoteEnvironment"`
}

type reDeleteMutationResponse struct {
	ReDeleteMutation reDeleteMutation `json:"deleteRemoteEnvironment"`
}

type remoteEnvironmentEvent struct {
	Type              string
	RemoteEnvironment remoteEnvironment
}

func TestRemoteEnvironmentMutations(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}
	c, err := graphql.New()
	require.NoError(t, err)

	reCli, _, err := client.NewREClientWithConfig()
	require.NoError(t, err)

	t.Log("Subscribe On Remote Environments")
	subscription := subscribeREEvent(c)
	defer subscription.Close()

	const fixName = "test-ui-api-re"
	fixedRE := createREStruct(fixName, "fix-desc1", map[string]string{"fix": "lab"})

	waitForREReady(fixName, reCli)

	t.Log("Create Remote Environment")
	resp, err := createRE(c, fixedRE)
	require.NoError(t, err)
	assert.Equal(t, *fixedRE, resp.ReCreateMutation)

	t.Log("Check Subscription Event")
	expectedEvent := createREEventStruct("ADD", fixedRE)
	event, err := readREEvent(subscription)
	assert.NoError(t, err)
	checkREEvent(t, expectedEvent, event)

	defer func() {
		t.Log("Delete Remote Environment")
		deleteResp, err := deleteRE(c, fixName)
		require.NoError(t, err)
		assert.Equal(t, fixName, deleteResp.ReDeleteMutation.Name)
	}()

	t.Log("Update Remote Environment")
	fixedRE = createREStruct(fixName, "desc2", map[string]string{"lab": "fix"})

	updateResp, err := updateRE(c, fixedRE)
	require.NoError(t, err)
	assert.Equal(t, *fixedRE, updateResp.ReUpdateMutation)
}

func createREStruct(name string, desc string, labels map[string]string) *remoteEnvironment {
	return &remoteEnvironment{
		Name:        name,
		Description: desc,
		Labels:      labels,
	}
}

func readREEvent(sub *graphql.Subscription) (remoteEnvironmentEvent, error) {
	type Response struct {
		RemoteEnvironmentEvent remoteEnvironmentEvent
	}
	var reEvent Response
	err := sub.Next(&reEvent, tester.DefaultSubscriptionTimeout)

	return reEvent.RemoteEnvironmentEvent, err
}

func subscribeREEvent(c *graphql.Client) *graphql.Subscription {
	query := fmt.Sprintf(`
			subscription {
				%s
			}
		`, reEventFields())
	req := graphql.NewRequest(query)
	return c.Subscribe(req)
}

func reEventFields() string {
	return fmt.Sprintf(`
        remoteEnvironmentEvent {
			type
			remoteEnvironment{
				%s
			}
        }
    `, reFields())
}

func reMutationOutputFields() string {
	return `
		name
		description
		labels
    `
}

func reFields() string {
	return `
		name
		description
		labels
		enabledInEnvironments
		status               
		services {
			id                 
			displayName        
			longDescription    
			providerDisplayName
			tags               
			entries {
				type       
				gatewayUrl 
				accessLabel
			}         
		}          
    `
}

func waitForREReady(environment string, reCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		re, err := reCli.ApplicationconnectorV1alpha1().RemoteEnvironments().Get(environment, metav1.GetOptions{})
		if err != nil || re == nil {
			return false, err
		}

		conditions := re.Status.Conditions
		for _, cond := range conditions {
			if cond.Type == `Ready` {
				return cond.Status == v1alpha1.ConditionTrue, nil
			}
		}

		return false, nil
	}, reReadyTimeout)
}

func checkREEvent(t *testing.T, expected, actual remoteEnvironmentEvent) {
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.RemoteEnvironment.Name, actual.RemoteEnvironment.Name)
}

func createREEventStruct(eventType string, re *remoteEnvironment) remoteEnvironmentEvent {
	return remoteEnvironmentEvent{
		Type:              eventType,
		RemoteEnvironment: *re,
	}
}

func createRE(c *graphql.Client, given *remoteEnvironment) (reCreateMutationResponse, error) {
	query := fmt.Sprintf(`
			mutation ($name: String!, $description: String!, $labels: Labels!) {
				createRemoteEnvironment(name: $name, description: $description, labels: $labels) {
					%s
				}
			}
	`, reMutationOutputFields())

	req := graphql.NewRequest(query)
	req.SetVar("name", given.Name)
	req.SetVar("description", given.Description)
	req.SetVar("labels", given.Labels)
	var response reCreateMutationResponse
	err := c.Do(req, &response)

	return response, err
}

func updateRE(c *graphql.Client, given *remoteEnvironment) (reUpdateMutationResponse, error) {
	query := fmt.Sprintf(`
			mutation ($name: String!, $description: String!, $labels: Labels!) {
				updateRemoteEnvironment(name: $name, description: $description, labels: $labels) {
					%s
				}
			}
	`, reMutationOutputFields())

	req := graphql.NewRequest(query)
	req.SetVar("name", given.Name)
	req.SetVar("description", given.Description)
	req.SetVar("labels", given.Labels)
	var response reUpdateMutationResponse
	err := c.Do(req, &response)

	return response, err
}

func deleteRE(c *graphql.Client, reName string) (reDeleteMutationResponse, error) {
	query := `
			mutation ($name: String!) {
				deleteRemoteEnvironment(name: $name) {
					name
				}
			}
	`
	req := graphql.NewRequest(query)
	req.SetVar("name", reName)
	var response reDeleteMutationResponse
	err := c.Do(req, &response)

	return response, err
}

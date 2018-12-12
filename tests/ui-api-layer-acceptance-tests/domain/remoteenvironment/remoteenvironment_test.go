// +build acceptance

package remoteenvironment

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	reReadyTimeout = time.Second * 45
)

type remoteEnvironmentMutationOutput struct {
	Name        string
	Description string
	Labels      map[string]string
}

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
	ReCreateMutation remoteEnvironmentMutationOutput `json:"createRemoteEnvironment"`
}

type reUpdateMutationResponse struct {
	ReUpdateMutation remoteEnvironmentMutationOutput `json:"updateRemoteEnvironment"`
}

type reDeleteMutationResponse struct {
	ReDeleteMutation remoteEnvironmentMutationOutput `json:"deleteRemoteEnvironment"`
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

	t.Log("Subscribe On Remote Environments")
	subscription := subscribeREEvent(c)
	defer subscription.Close()

	const fixName = "test-ui-api-re"
	fixedRE := fixRE(fixName, "fix-desc1", map[string]string{"fix": "lab"})

	t.Log("Create Remote Environment")
	resp, err := createREMutation(c, fixedRE)
	require.NoError(t, err)
	checkREMutationOutput(t, fixedRE, resp.ReCreateMutation)

	t.Log("Check Subscription Event")
	expectedEvent := fixREEvent("ADD", fixedRE)
	event, err := readREEvent(subscription)
	assert.NoError(t, err)
	checkREEvent(t, expectedEvent, event)

	defer func() {
		t.Log("Delete Remote Environment")
		deleteResp, err := deleteREMutation(c, fixName)
		require.NoError(t, err)
		assert.Equal(t, fixName, deleteResp.ReDeleteMutation.Name)
	}()

	t.Log("Update Remote Environment")
	fixedRE = fixRE(fixName, "desc2", map[string]string{"lab": "fix"})

	updateResp, err := updateREMutation(c, fixedRE)
	require.NoError(t, err)
	checkREMutationOutput(t, fixedRE, updateResp.ReUpdateMutation)
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

func checkREEvent(t *testing.T, expected, actual remoteEnvironmentEvent) {
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.RemoteEnvironment.Name, actual.RemoteEnvironment.Name)
}

func checkREMutationOutput(t *testing.T, re *remoteEnvironment, reMutation remoteEnvironmentMutationOutput) {
	assert.Equal(t, re.Name, reMutation.Name)
	assert.Equal(t, re.Description, reMutation.Description)
	assert.Equal(t, re.Labels, reMutation.Labels)
}

func createREMutation(c *graphql.Client, given *remoteEnvironment) (reCreateMutationResponse, error) {
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

func updateREMutation(c *graphql.Client, given *remoteEnvironment) (reUpdateMutationResponse, error) {
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

func deleteREMutation(c *graphql.Client, reName string) (reDeleteMutationResponse, error) {
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

func fixRE(name string, desc string, labels map[string]string) *remoteEnvironment {
	return &remoteEnvironment{
		Name:        name,
		Description: desc,
		Labels:      labels,
	}
}

func fixREEvent(eventType string, re *remoteEnvironment) remoteEnvironmentEvent {
	return remoteEnvironmentEvent{
		Type:              eventType,
		RemoteEnvironment: *re,
	}
}

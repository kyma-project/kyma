// +build acceptance

package remoteenvironment

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"

	clientset "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/k8s"
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

type RemoteEnvironment struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

type reDeleteMutation struct {
	Name string `json:"name"`
}

type reCreateMutationResponse struct {
	ReCreateMutation RemoteEnvironment `json:"createRemoteEnvironment"`
}

type reUpdateMutationResponse struct {
	ReUpdateMutation RemoteEnvironment `json:"updateRemoteEnvironment"`
}

type reDeleteMutationResponse struct {
	ReDeleteMutation reDeleteMutation `json:"deleteRemoteEnvironment"`
}

type RemoteEnvironmentEvent struct {
	Type              string
	RemoteEnvironment RemoteEnvironment
}

func TestRemoteEnvironmentMutations(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}
	c, err := graphql.New()
	require.NoError(t, err)

	reCli, _, err := k8s.NewREClientWithConfig()
	require.NoError(t, err)

	t.Log("Remote Environment Events Subscription")
	subscription := subscribeREEvent(c)
	defer subscription.Close()

	const fixName = "test-ui-api-re"
	fixREResponse := RemoteEnvironment{
		Name:        fixName,
		Description: "fix-desc1",
		Labels: map[string]string{
			"fix": "lab",
		},
	}

	waitForREReady(fixName, reCli)

	t.Log("Create RemoteEnvironment")
	resp, err := createRE(c, fixREResponse)
	require.NoError(t, err)
	assert.Equal(t, fixREResponse, resp.ReCreateMutation)

	t.Log("Check Remote Environment Event")
	expectedEvent := remoteEnvironmentEvent("ADD", fixREResponse)
	event, err := readRemoteEnvironmentEvent(subscription)
	assert.NoError(t, err)
	checkREEvent(t, expectedEvent, event)

	defer func() {
		t.Log("Delete RemoteEnvironment")
		deleteResp, err := deleteRE(c, fixName)
		require.NoError(t, err)
		assert.Equal(t, fixName, deleteResp.ReDeleteMutation.Name)
	}()

	t.Log("Update RemoteEnvironment")
	fixREResponse.Description = "desc2"
	fixREResponse.Labels = map[string]string{
		"lab": "fix",
	}
	updateResp, err := updateRE(c, fixREResponse)
	require.NoError(t, err)
	assert.Equal(t, fixREResponse, updateResp.ReUpdateMutation)
}

func readRemoteEnvironmentEvent(sub *graphql.Subscription) (RemoteEnvironmentEvent, error) {
	type Response struct {
		RemoteEnvironmentEvent RemoteEnvironmentEvent
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
		`, remoteEnvironmentEventFields())
	req := graphql.NewRequest(query)
	return c.Subscribe(req)
}

func remoteEnvironmentEventFields() string {
	return `
        remoteEnvironmentEvent {
			type
    		remoteEnvironment{
				name
				description
				labels
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

func checkREEvent(t *testing.T, expected, actual RemoteEnvironmentEvent) {
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.RemoteEnvironment.Name, actual.RemoteEnvironment.Name)
}

func remoteEnvironmentEvent(eventType string, re RemoteEnvironment) RemoteEnvironmentEvent {
	return RemoteEnvironmentEvent{
		Type:              eventType,
		RemoteEnvironment: re,
	}
}

func createRE(c *graphql.Client, given RemoteEnvironment) (reCreateMutationResponse, error) {
	query := `
			mutation ($name: String!, $description: String!, $labels: Labels!) {
				createRemoteEnvironment(name: $name, description: $description, labels: $labels) {
					name
					description
					labels
				}
			}
	`
	req := graphql.NewRequest(query)
	req.SetVar("name", given.Name)
	req.SetVar("description", given.Description)
	req.SetVar("labels", given.Labels)
	var response reCreateMutationResponse
	err := c.Do(req, &response)

	return response, err
}

func updateRE(c *graphql.Client, given RemoteEnvironment) (reUpdateMutationResponse, error) {
	query := `
			mutation ($name: String!, $description: String!, $labels: Labels!) {
				updateRemoteEnvironment(name: $name, description: $description, labels: $labels) {
					name
					description
					labels
				}
			}
	`
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

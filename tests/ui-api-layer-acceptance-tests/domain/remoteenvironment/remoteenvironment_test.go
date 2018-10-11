// +build acceptance

package remoteenvironment

import (
	"testing"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type reMutationData struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

type reDeleteMutation struct {
	Name string `json:"name"`
}

type reCreateMutationResponse struct {
	ReCreateMutation reMutationData `json:"createRemoteEnvironment"`
}

type reUpdateMutationResponse struct {
	ReUpdateMutation reMutationData `json:"updateRemoteEnvironment"`
}

type reDeleteMutationResponse struct {
	ReDeleteMutation reDeleteMutation `json:"deleteRemoteEnvironment"`
}

func TestRemoteEnvironmentMutations(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}
	c, err := graphql.New()
	require.NoError(t, err)

	const fixName = "test-ui-api-re"
	fixResponse := reMutationData{
		Name:        fixName,
		Description: "fix-desc1",
		Labels: map[string]string{
			"fix": "lab",
		},
	}
	t.Log("Create RemoteEnvironment")
	resp, err := createRE(c, fixResponse)
	require.NoError(t, err)
	assert.Equal(t, fixResponse, resp.ReCreateMutation)

	defer func() {
		t.Log("Delete RemoteEnvironment")
		deleteResp, err := deleteRE(c, fixName)
		require.NoError(t, err)
		assert.Equal(t, fixName, deleteResp.ReDeleteMutation.Name)
	}()

	t.Log("Update RemoteEnvironment")
	fixResponse.Description = "desc2"
	fixResponse.Labels = map[string]string{
		"lab": "fix",
	}
	updateResp, err := updateRE(c, fixResponse)
	require.NoError(t, err)
	assert.Equal(t, fixResponse, updateResp.ReUpdateMutation)
}

func createRE(c *graphql.Client, given reMutationData) (reCreateMutationResponse, error) {
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

func updateRE(c *graphql.Client, given reMutationData) (reUpdateMutationResponse, error) {
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

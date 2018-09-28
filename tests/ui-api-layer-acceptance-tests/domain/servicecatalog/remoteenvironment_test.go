package servicecatalog

import (
	"testing"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/stretchr/testify/require"
	"github.com/magiconair/properties/assert"
)

type reMutationResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

type deleteReResponse struct {
	Name string `json:"name"`
}

func TestRemoteEnvironmentMutations(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}
	c, err := graphql.New()
	require.NoError(t, err)

	const fixName = "test-re1"
	expectedResponse := reMutationResponse{
		Name:        fixName,
		Description: "fix-desc1",
		Labels: map[string]string{
			"fix": "lab",
		},
	}
	t.Log("Create RemoteEnvironment")
	resp, err := createRE(c, expectedResponse)
	require.NoError(t, err)
	assert.Equal(t, resp, expectedResponse)

	defer func() {
		t.Log("Delete RemoteEnvironment")
		reName, err := deleteRE(c, fixName)
		require.NoError(t, err)
		assert.Equal(t, reName.Name, fixName)
	}()

	t.Log("Update RemoteEnvironment")
	expectedResponse.Description = "desc2"
	expectedResponse.Labels = map[string]string{
		"lab": "fix",
	}
	resp, err = updateRE(c, expectedResponse)
	require.NoError(t, err)
	assert.Equal(t, resp, expectedResponse)
}

func createRE(c *graphql.Client, givenResponse reMutationResponse) (reMutationResponse, error) {
	query := `
			mutation ($name: String!, $description: String!, $labels: JSON!) {
				createRemoteEnvironment(name: $name, description: $description, labels: $labels) {
					name
					description
					labels
				}
			}
	`
	req := graphql.NewRequest(query)
	req.SetVar("name", givenResponse.Name)
	req.SetVar("description", givenResponse.Description)
	req.SetVar("labels", givenResponse.Labels)
	var response reMutationResponse
	err := c.Do(req, &response)

	return response, err
}

func updateRE(c *graphql.Client, givenResponse reMutationResponse) (reMutationResponse, error) {
	query := `
			mutation ($name: String!, $description: String!, $labels: JSON!) {
				updateRemoteEnvironment(name: $name, description: $description, labels: $labels) {
					name
					description
					labels
				}
			}
	`
	req := graphql.NewRequest(query)
	req.SetVar("name", givenResponse.Name)
	req.SetVar("description", givenResponse.Description)
	req.SetVar("labels", givenResponse.Labels)
	var response reMutationResponse
	err := c.Do(req, &response)

	return response, err
}

func deleteRE(c *graphql.Client, reName string) (deleteReResponse, error) {
	query := `
			mutation ($name: String!) {
				deleteRemoteEnvironment(name: $name) {
					name
				}
			}
	`
	req := graphql.NewRequest(query)
	req.SetVar("name", reName)
	var response deleteReResponse
	err := c.Do(req, &response)

	return response, err
}

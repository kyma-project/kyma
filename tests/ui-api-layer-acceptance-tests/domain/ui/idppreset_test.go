// +build acceptance

package ui

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	IDPPresetReadyTimeout    = time.Second * 45
	IDPPresetDeletionTimeout = time.Second * 15
)

type IDPPreset struct {
	Name    string
	Issuer  string
	JwksUri string
}

type idpPresetCreateMutationResponse struct {
	CreateIDPPreset IDPPreset
}

type idpPresetQueryResponse struct {
	IDPPreset IDPPreset
}

type idpPresetsQueryResponse struct {
	IDPPresets []IDPPreset
}

func TestIDPPresetQueriesAndMutations(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	// client, _, err := k8s.NewIDPPresetClientWithConfig()
	// require.NoError(t, err)

	expectedResource := idpPreset("test-name7", "test-issuer", "https://test-jwksUri")
	resourceDetailsQuery := idpPresetDetailsFields()

	t.Log("Create IDPPreset")
	createRes, err := createIDPPreset(c, resourceDetailsQuery, expectedResource)

	require.NoError(t, err)
	checkIDPPreset(t, expectedResource, createRes.CreateIDPPreset)

	t.Log("Query Single Resource")
	res, err := querySingleIDPPreset(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	checkIDPPreset(t, expectedResource, res.IDPPreset)

	t.Log("Query Multiple Resources")
	multipleRes, err := queryMultipleIDPPresets(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	assertIDPPresetExistsAndEqual(t, expectedResource, multipleRes.IDPPresets)

}

func idpPreset(name string, issuer string, jwksUri string) IDPPreset {
	return IDPPreset{
		Name:    name,
		Issuer:  issuer,
		JwksUri: jwksUri,
	}
}

func idpPresetDetailsFields() string {
	return `
		name
		issuer
		jwksUri
	`
}

func createIDPPreset(c *graphql.Client, resourceDetailsQuery string, expectedResource IDPPreset) (idpPresetCreateMutationResponse, error) {
	query := fmt.Sprintf(`
		mutation ($name: String!, $issuer: String!, $jwksUri: String!) {
			createIDPPreset(
				name: $name,
				issuer: $issuer,
				jwksUri: $jwksUri,
			) {
				%s
			}
		}
	`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)
	req.SetVar("issuer", expectedResource.Issuer)
	req.SetVar("jwksUri", expectedResource.JwksUri)

	var res idpPresetCreateMutationResponse
	err := c.Do(req, &res)

	return res, err
}

func querySingleIDPPreset(c *graphql.Client, resourceDetailsQuery string, expectedResource IDPPreset) (idpPresetQueryResponse, error) {
	req := singleResourceQueryRequest(resourceDetailsQuery, expectedResource)

	var res idpPresetQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func singleResourceQueryRequest(resourceDetailsQuery string, expectedResource IDPPreset) *graphql.Request {
	query := fmt.Sprintf(`
			query ($name: String!) {
				IDPPreset(name: $name) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)

	return req
}

func queryMultipleIDPPresets(c *graphql.Client, resourceDetailsQuery string, expectedResource IDPPreset) (idpPresetsQueryResponse, error) {
	req := multipleResourcesQueryRequest(resourceDetailsQuery, expectedResource)

	var res idpPresetsQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func multipleResourcesQueryRequest(resourceDetailsQuery string, expectedResource IDPPreset) *graphql.Request {
	query := fmt.Sprintf(`
			query () {
				IDPPresets() {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)

	return req
}

func assertIDPPresetExistsAndEqual(t *testing.T, expectedElement IDPPreset, arr []IDPPreset) {

	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkIDPPreset(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

func checkIDPPreset(t *testing.T, expected, actual IDPPreset) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Issuer, actual.Issuer)
	assert.Equal(t, expected.JwksUri, actual.JwksUri)
}

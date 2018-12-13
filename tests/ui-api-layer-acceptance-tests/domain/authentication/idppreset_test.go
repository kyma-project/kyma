// +build acceptance

package authentication

import (
	"fmt"
	"testing"
	"time"

	idpClientset "github.com/kyma-project/kyma/components/idppreset/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/client"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const idpPresetMutationTimeout = time.Second * 15

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
type idpPresetDeleteMutationResponse struct {
	DeleteIDPPreset IDPPreset
}

func TestIDPPresetQueriesAndMutations(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}

	c, err := graphql.New()
	require.NoError(t, err)

	client, _, err := client.NewIDPPresetClientWithConfig()
	require.NoError(t, err)

	expectedResource := idpPreset("test-name", "test-issuer", "https://test-jwksUri")
	resourceDetailsQuery := idpPresetDetailsFields()

	t.Log("Create IDP Preset")
	createRes, err := createIDPPreset(c, resourceDetailsQuery, expectedResource)

	require.NoError(t, err)
	checkIDPPreset(t, expectedResource, createRes.CreateIDPPreset)

	t.Log("Wait For IDP Preset Creation")
	err = waitForIDPPresetCreation(expectedResource.Name, client)
	assert.NoError(t, err)

	t.Log("Query Single Resource")
	res, err := querySingleIDPPreset(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	checkIDPPreset(t, expectedResource, res.IDPPreset)

	t.Log("Query Multiple Resources")
	multipleRes, err := queryMultipleIDPPresets(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	checkIDPPresets(t, expectedResource, multipleRes.IDPPresets)

	t.Log("Delete IDP Preset")
	deleteRes, err := deleteIDPPreset(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	checkIDPPreset(t, expectedResource, deleteRes.DeleteIDPPreset)

	t.Log("Wait For IDP Preset Deletion")
	err = waitForIDPPresetDeletion(expectedResource.Name, client)
	assert.NoError(t, err)
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

func waitForIDPPresetCreation(name string, client *idpClientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		_, err := client.AuthenticationV1alpha1().IDPPresets().Get(name, metav1.GetOptions{})

		if err != nil {
			return false, err
		}

		return true, nil
	}, idpPresetMutationTimeout)
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
			query {
				IDPPresets {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)

	return req
}

func deleteIDPPreset(c *graphql.Client, resourceDetailsQuery string, expectedResource IDPPreset) (idpPresetDeleteMutationResponse, error) {
	query := fmt.Sprintf(`
			mutation ($name: String!) {
				deleteIDPPreset(name: $name) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)

	var res idpPresetDeleteMutationResponse
	err := c.Do(req, &res)

	return res, err
}

func waitForIDPPresetDeletion(name string, client *idpClientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		_, err := client.AuthenticationV1alpha1().IDPPresets().Get(name, metav1.GetOptions{})

		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		return false, nil
	}, idpPresetMutationTimeout)
}

func checkIDPPresets(t *testing.T, expectedElement IDPPreset, arr []IDPPreset) {

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

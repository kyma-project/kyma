// +build acceptance

package ui

import (
	"fmt"
	"testing"
	"time"

	idpClientset "github.com/kyma-project/kyma/components/idppreset/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/k8s"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestIDPPresetQueriesAndMutations(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	client, _, err := k8s.NewIDPPresetClientWithConfig()
	require.NoError(t, err)

	expectedResource := idpPreset("test-name7", "test-issuer", "https://test-jwksUri")
	resourceDetailsQuery := idpPresetDetailsFields()

	t.Log("Create IDPPreset")
	createRes, err := createIDPPreset(c, resourceDetailsQuery, expectedResource)

	require.NoError(t, err)
	checkIDPPreset(t, expectedResource, createRes.CreateIDPPreset)

	t.Log("Wait For IDPPreset Ready")
	err = waitForIDPPresetReady(expectedResource.Name, expectedResource.Issuer, expectedResource.JwksUri, client)
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

func checkIDPPreset(t *testing.T, expected, actual IDPPreset) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Issuer, actual.Issuer)
	assert.Equal(t, expected.JwksUri, actual.JwksUri)
}

func waitForIDPPresetReady(name string, issuer string, jwksUri string, client *idpClientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		idp, err := client.UiV1alpha1().IDPPresets().Get(name, metav1.GetOptions{})
		if err != nil || idp == nil {
			return false, err
		}

		arr := idp.Spec.Issuer
		if arr == issuer {
			return true, nil
		}

		return false, nil
	}, IDPPresetReadyTimeout)
}

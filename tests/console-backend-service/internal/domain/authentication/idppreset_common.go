// +build acceptance

package authentication

import (
	"fmt"
	"testing"
	"time"

	idpClientset "github.com/kyma-project/kyma/components/idppreset/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/waiter"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IDPPreset struct {
	Name    string
	Issuer  string
	JwksUri string
}

func idpPreset(name string, issuer string, jwksUri string) IDPPreset {
	return IDPPreset{
		Name:    name,
		Issuer:  issuer,
		JwksUri: jwksUri,
	}
}

const idpPresetMutationTimeout = time.Second * 15

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

func idpPresetDetailsFields() string {
	return `
		name
		issuer
		jwksUri
	`
}

func fixCreateIDPPresetRequest(resourceDetailsQuery string, idpPreset IDPPreset) *graphql.Request {
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
	req.SetVar("name", idpPreset.Name)
	req.SetVar("issuer", idpPreset.Issuer)
	req.SetVar("jwksUri", idpPreset.JwksUri)

	return req
}

func createIDPPreset(c *graphql.Client, resourceDetailsQuery string, expectedResource IDPPreset) (idpPresetCreateMutationResponse, error) {
	req := fixCreateIDPPresetRequest(resourceDetailsQuery, expectedResource)

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

func fixDeleteIDPPresetRequest(resourceDetailsQuery string, expectedResource IDPPreset) *graphql.Request {
	query := fmt.Sprintf(`
			mutation ($name: String!) {
				deleteIDPPreset(name: $name) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)

	return req
}

func deleteIDPPreset(c *graphql.Client, resourceDetailsQuery string, expectedResource IDPPreset) (idpPresetDeleteMutationResponse, error) {
	req := fixDeleteIDPPresetRequest(resourceDetailsQuery, expectedResource)

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

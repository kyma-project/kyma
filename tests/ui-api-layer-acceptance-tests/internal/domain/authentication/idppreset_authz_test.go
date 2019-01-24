// +build acceptance

package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/graphql"
	"github.com/stretchr/testify/require"
)

const (
	RBACError string = "graphql: access denied"
	AuthError string = "graphql: server returned a non-200 status code: 401"
)

func TestAuthenticationAnonymousUser(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

	c := gqlClientForUser(t, graphql.NoUser)

	testResource := idpPreset("test-name", "test-issuer", "https://test-jwksUri")
	resourceDetailsQuery := idpPresetDetailsFields()

	// should not be able to do any request  - should receive 401 Unauthorized
	t.Log("Query Single IDP Preset")
	res, err := querySingleIDPPreset(c, resourceDetailsQuery, testResource)

	assert.EqualError(t, err, AuthError)
	assert.Equal(t, IDPPreset{}, res.IDPPreset)

}

func TestAuthorizationForNoRightsUser(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

	c := gqlClientForUser(t, graphql.NoRightsUser)

	testResource := idpPreset("test-name", "test-issuer", "https://test-jwksUri")
	resourceDetailsQuery := idpPresetDetailsFields()

	// should not be able to get idppresets
	t.Log("Query Single IDP Preset")
	queryRes, err := querySingleIDPPreset(c, resourceDetailsQuery, testResource)

	assert.Equal(t, IDPPreset{}, queryRes.IDPPreset)
	assert.EqualError(t, err, RBACError)

	// should not be able to create idppreset
	t.Log("Create IDP Preset")
	createRes, err := createIDPPreset(c, resourceDetailsQuery, testResource)

	assert.Equal(t, IDPPreset{}, createRes.CreateIDPPreset)
	assert.EqualError(t, err, RBACError)

	// should not be able to delete idppreset
	t.Log("Delete IDP Preset")
	deleteRes, err := deleteIDPPreset(c, resourceDetailsQuery, testResource)

	assert.Equal(t, IDPPreset{}, deleteRes.DeleteIDPPreset)
	assert.EqualError(t, err, RBACError)
}

func TestAuthorizationForReadOnlyUser(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

	c := gqlClientForUser(t, graphql.ReadOnlyUser)

	testResource := idpPreset("test-name", "test-issuer", "https://test-jwksUri")
	resourceDetailsQuery := idpPresetDetailsFields()

	// should be able to get idppresets
	t.Log("Query Single IDP Preset")
	queryRes, err := querySingleIDPPreset(c, resourceDetailsQuery, testResource)

	assert.Equal(t, IDPPreset{}, queryRes.IDPPreset)
	assert.Nil(t, err)

	// should not be able to create idppreset
	t.Log("Create IDP Preset")
	createRes, err := createIDPPreset(c, resourceDetailsQuery, testResource)

	assert.Equal(t, IDPPreset{}, createRes.CreateIDPPreset)
	assert.EqualError(t, err, RBACError)

	// should not be able to delete idppreset
	t.Log("Delete IDP Preset")
	deleteRes, err := deleteIDPPreset(c, resourceDetailsQuery, testResource)

	assert.Equal(t, IDPPreset{}, deleteRes.DeleteIDPPreset)
	assert.EqualError(t, err, RBACError)
}

func TestAuthorizationForReadWriteUser(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

	c := gqlClientForUser(t, graphql.AdminUser)

	testResource := idpPreset("test-name", "test-issuer", "https://test-jwksUri")
	resourceDetailsQuery := idpPresetDetailsFields()

	// should be able to get idppresets
	t.Log("Query Single IDP Preset")
	queryRes, err := querySingleIDPPreset(c, resourceDetailsQuery, testResource)

	checkIDPPreset(t, IDPPreset{}, queryRes.IDPPreset)
	assert.Nil(t, err)

	// should be able to create idppreset
	t.Log("Create IDP Preset")
	createRes, err := createIDPPreset(c, resourceDetailsQuery, testResource)

	checkIDPPreset(t, testResource, createRes.CreateIDPPreset)
	assert.Nil(t, err)

	// should be able to delete idppreset
	t.Log("Delete IDP Preset")
	deleteRes, err := deleteIDPPreset(c, resourceDetailsQuery, testResource)

	checkIDPPreset(t, testResource, deleteRes.DeleteIDPPreset)
	assert.Nil(t, err)
}

func gqlClientForUser(t *testing.T, user graphql.User) *graphql.Client {
	c, err := graphql.New()
	require.NoError(t, err)

	err = c.ChangeUser(user)
	require.NoError(t, err)

	return c
}

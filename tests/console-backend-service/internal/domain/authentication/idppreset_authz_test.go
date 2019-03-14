// +build acceptance

package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/dex"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/module"
	"github.com/stretchr/testify/require"
)

const (
	RBACError           string = "graphql: access denied"
	AuthError           string = "graphql: server returned a non-200 status code: 401"
	ModuleDisabledError string = "graphql: MODULE_DISABLED: The authentication module is disabled."
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

	t.Run("User with no rights", func(t *testing.T) {

		c := gqlClientForUser(t, graphql.NoRightsUser)

		testResource := idpPreset("test-name", "test-issuer", "https://test-jwksUri")
		resourceDetailsQuery := idpPresetDetailsFields()

		t.Run("should not be able to get idppresets", func(t *testing.T) {
			queryRes, err := querySingleIDPPreset(c, resourceDetailsQuery, testResource)

			assert.Equal(t, IDPPreset{}, queryRes.IDPPreset)
			assert.EqualError(t, err, RBACError)
		})

		t.Run("should not be able to create idppreset", func(t *testing.T) {
			createRes, err := createIDPPreset(c, resourceDetailsQuery, testResource)

			assert.Equal(t, IDPPreset{}, createRes.CreateIDPPreset)
			assert.EqualError(t, err, RBACError)
		})

		t.Run("should not be able to delete idppreset", func(t *testing.T) {
			deleteRes, err := deleteIDPPreset(c, resourceDetailsQuery, testResource)

			assert.Equal(t, IDPPreset{}, deleteRes.DeleteIDPPreset)
			assert.EqualError(t, err, RBACError)
		})

	})
}

func TestAuthorizationForReadOnlyUser(t *testing.T) {

	dex.SkipTestIfSCIEnabled(t)

	t.Run("Read only user", func(t *testing.T) {

		c := gqlClientForUser(t, graphql.ReadOnlyUser)

		moduleEnabled, err := module.IsEnabled(ModuleName, c)
		assert.Nil(t, err)

		testResource := idpPreset("test-name", "test-issuer", "https://test-jwksUri")
		resourceDetailsQuery := idpPresetDetailsFields()

		t.Run("should be able to get idppresets", func(t *testing.T) {
			queryRes, err := querySingleIDPPreset(c, resourceDetailsQuery, testResource)

			assert.Equal(t, IDPPreset{}, queryRes.IDPPreset)
			if moduleEnabled {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, ModuleDisabledError)
			}
		})

		t.Run("should not be able to create idppreset", func(t *testing.T) {
			createRes, err := createIDPPreset(c, resourceDetailsQuery, testResource)

			assert.Equal(t, IDPPreset{}, createRes.CreateIDPPreset)
			assert.EqualError(t, err, RBACError)
		})

		t.Run("should not be able to delete idppreset", func(t *testing.T) {
			deleteRes, err := deleteIDPPreset(c, resourceDetailsQuery, testResource)

			assert.Equal(t, IDPPreset{}, deleteRes.DeleteIDPPreset)
			assert.EqualError(t, err, RBACError)
		})
	})
}

func TestAuthorizationForReadWriteUser(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

	t.Run("Read write user", func(t *testing.T) {
		c := gqlClientForUser(t, graphql.AdminUser)

		moduleEnabled, err := module.IsEnabled(ModuleName, c)
		assert.Nil(t, err)

		testResource := idpPreset("test-name", "test-issuer", "https://test-jwksUri")
		resourceDetailsQuery := idpPresetDetailsFields()

		t.Run("should be able to get idppresets", func(t *testing.T) {
			queryRes, err := querySingleIDPPreset(c, resourceDetailsQuery, testResource)

			checkIDPPreset(t, IDPPreset{}, queryRes.IDPPreset)
			if moduleEnabled {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, ModuleDisabledError)
			}
		})

		t.Run("should be able to create idppreset", func(t *testing.T) {
			createRes, err := createIDPPreset(c, resourceDetailsQuery, testResource)

			if moduleEnabled {
				checkIDPPreset(t, testResource, createRes.CreateIDPPreset)
				assert.Nil(t, err)
			} else {
				checkIDPPreset(t, IDPPreset{}, createRes.CreateIDPPreset)
				assert.EqualError(t, err, ModuleDisabledError)
			}
		})

		t.Run("should be able to delete idppreset", func(t *testing.T) {
			deleteRes, err := deleteIDPPreset(c, resourceDetailsQuery, testResource)

			if moduleEnabled {
				checkIDPPreset(t, testResource, deleteRes.DeleteIDPPreset)
				assert.Nil(t, err)
			} else {
				checkIDPPreset(t, IDPPreset{}, deleteRes.DeleteIDPPreset)
				assert.EqualError(t, err, ModuleDisabledError)
			}
		})

	})

}

func gqlClientForUser(t *testing.T, user graphql.User) *graphql.Client {
	c, err := graphql.New()
	require.NoError(t, err)

	err = c.ChangeUser(user)
	require.NoError(t, err)

	return c
}

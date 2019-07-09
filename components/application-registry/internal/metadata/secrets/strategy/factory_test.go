package strategy

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
	"github.com/stretchr/testify/assert"
)

func TestFactory_NewSecretModificationStrategy(t *testing.T) {

	testCases := []struct {
		credentials *model.CredentialsWithCSRF
		strategy    ModificationStrategy
	}{
		{
			credentials: basicCredentials,
			strategy:    &basicAuth{},
		},
		{
			credentials: oauthCredentials,
			strategy:    &oauth{},
		},
		{
			credentials: certGenCredentials,
			strategy:    &certificateGen{},
		},
	}

	t.Run("should create new modification strategy", func(t *testing.T) {
		// given
		factory := &factory{}

		for _, test := range testCases {
			// when
			strategy, err := factory.NewSecretModificationStrategy(test.credentials)

			// then
			require.NoError(t, err)
			assert.IsType(t, test.strategy, strategy)
		}
	})

	t.Run("should return error when no credentials provided", func(t *testing.T) {
		// given
		factory := &factory{}

		// when
		_, err := factory.NewSecretModificationStrategy(&model.CredentialsWithCSRF{})

		// then
		require.Error(t, err)

	})

}

func TestFactory_NewSecretAccessStrategy(t *testing.T) {
	testCases := []struct {
		credentials *applications.Credentials
		strategy    AccessStrategy
	}{
		{
			credentials: &applications.Credentials{Type: applications.CredentialsBasicType},
			strategy:    &basicAuth{},
		},
		{
			credentials: &applications.Credentials{Type: applications.CredentialsOAuthType},
			strategy:    &oauth{},
		},
		{
			credentials: &applications.Credentials{Type: applications.CredentialsCertificateGenType},
			strategy:    &certificateGen{},
		},
	}

	t.Run("should create new access strategy", func(t *testing.T) {
		// given
		factory := &factory{}

		for _, test := range testCases {
			// when
			strategy, err := factory.NewSecretAccessStrategy(test.credentials)

			// then
			require.NoError(t, err)
			assert.IsType(t, test.strategy, strategy)
		}
	})

	t.Run("should return error when no credentials provided", func(t *testing.T) {
		// given
		factory := &factory{}

		// when
		_, err := factory.NewSecretAccessStrategy(&applications.Credentials{})

		// then
		require.Error(t, err)
	})
}

package strategy

import (
	"testing"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory_NewSecretModificationStrategy(t *testing.T) {

	testCases := []struct {
		credentials *model.Credentials
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
		_, err := factory.NewSecretModificationStrategy(&model.Credentials{})

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

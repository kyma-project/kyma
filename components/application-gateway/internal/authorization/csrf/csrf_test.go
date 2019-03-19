package csrf

import (
	"testing"

	authmocks "github.com/kyma-project/kyma/components/application-gateway/internal/authorization/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestTokenEndpointURL = "myapp.com/csrf/token"
	noURL                = ""
)

func TestStrategyFactory_Create(t *testing.T) {

	// given
	factory := NewTokenStrategyFactory(nil)
	authStrategy := &authmocks.Strategy{}

	t.Run("Should create strategy if the CSRF token endpoint URL has been provided", func(t *testing.T) {

		// when
		tokenStrategy := factory.Create(authStrategy, TestTokenEndpointURL)

		// then
		require.NotNil(t, tokenStrategy)
		assert.IsType(t, &strategy{}, tokenStrategy)
	})

	t.Run("Should create noTokenStrategy if the CSRF token has not been provided", func(t *testing.T) {

		// when
		tokenStrategy := factory.Create(authStrategy, noURL)

		// then
		require.NotNil(t, tokenStrategy)
		assert.IsType(t, &noTokenStrategy{}, tokenStrategy)
	})
}

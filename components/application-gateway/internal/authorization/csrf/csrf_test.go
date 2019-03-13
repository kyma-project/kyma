package csrf

import (
	"github.com/kyma-project/kyma/components/application-gateway/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/internal/authorization/csrf/mocks"
	authmocks "github.com/kyma-project/kyma/components/application-gateway/internal/authorization/mocks"
	"github.com/kyma-project/kyma/components/application-gateway/internal/httpconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
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

func TestStrategy_AddCSRFToken(t *testing.T) {

	// given
	c := &mocks.Client{}
	s := &strategy{nil, TestTokenEndpointURL, c}

	req := &http.Request{}

	cachedItem := &Response{
		csrfToken: cachedToken,
		cookies: []*http.Cookie{
			{Name: cachedCookieName},
		},
	}

	t.Run("Should set CSRF header and copy all cookies into the request if it is possible to fetch the CSRF token", func(t *testing.T) {

		c.On("GetTokenEndpointResponse").Return(cachedItem, nil)

		// when
		err := s.AddCSRFToken(req)

		//then
		require.Nil(t, err)
		assert.Equal(t, cachedToken, req.Header.Get(httpconsts.HeaderCSRFToken))
		assert.Equal(t, cachedCookieName, req.Cookies()[0].Name)

	})

	t.Run("Should return error if it is not possible to fetch the CSRF token", func(t *testing.T) {

		c.On("GetTokenEndpointResponse").Return(nil, apperrors.NotFound("error"))

		//when
		err := s.AddCSRFToken(req)

		//then
		require.NotNil(t, err)
	})
}

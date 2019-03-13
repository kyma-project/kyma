package csrf_test

import (
	"github.com/kyma-project/kyma/components/application-gateway/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/internal/authorization/csrf"
	"github.com/kyma-project/kyma/components/application-gateway/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-gateway/internal/authorization/csrf/mocks"
	authmocks "github.com/kyma-project/kyma/components/application-gateway/internal/authorization/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

const (
	cachedToken = "someToken"
	cachedCookieName = "someCookieName"
	testCSRFTokenEndpointURL = "app.com/token"
)

func TestStrategy_AddCSRFToken(t *testing.T) {

	authStrategy := &authmocks.Strategy{}

	req := &http.Request{
		Header: make(map[string][]string),
	}


	t.Run("Should set CSRF header and copy all Cookies into the request if it is possible to fetch the CSRF token", func(t *testing.T) {

		// given
		c := &mocks.Client{}
		sf := csrf.NewTokenStrategyFactory(c)

		s := sf.Create(authStrategy, testCSRFTokenEndpointURL)

		cachedItem := &csrf.Response{
			CSRFToken: cachedToken,
			Cookies: []*http.Cookie{
				{Name: cachedCookieName},
			},
		}

		c.On("GetTokenEndpointResponse", testCSRFTokenEndpointURL, authStrategy).Return(cachedItem, nil)

		// when
		err := s.AddCSRFToken(req)

		//then
		require.Nil(t, err)
		assert.Equal(t, cachedToken, req.Header.Get(httpconsts.HeaderCSRFToken))
		assert.Equal(t, cachedCookieName, req.Cookies()[0].Name)

	})

	t.Run("Should return error if it is not possible to fetch the CSRF token", func(t *testing.T) {

		// given
		c := &mocks.Client{}
		sf := csrf.NewTokenStrategyFactory(c)

		s := sf.Create(authStrategy, testCSRFTokenEndpointURL)

		c.On("GetTokenEndpointResponse", testCSRFTokenEndpointURL, authStrategy).Return(nil, apperrors.NotFound("error"))

		//when
		err := s.AddCSRFToken(req)

		//then
		require.NotNil(t, err)
	})
}
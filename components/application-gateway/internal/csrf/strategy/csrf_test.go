package strategy

import (
	"github.com/kyma-project/kyma/components/application-gateway/internal/csrf/mocks"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	authmocks "github.com/kyma-project/kyma/components/application-gateway/pkg/authorization/mocks"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/httpconsts"
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

const (
	cachedToken              = "someToken"
	cachedCookieName         = "someCookieName"
	testCSRFTokenEndpointURL = "app.com/token"
)

func TestStrategy_AddCSRFToken(t *testing.T) {

	t.Run("In case CSRF is enabled", func(t *testing.T) {

		authStrategy := &authmocks.Strategy{}

		t.Run("Should set CSRF header and copy all Cookies into the request if it is possible to fetch the CSRF token", func(t *testing.T) {

			// given
			req := getNewEmptyRequest()

			c := &mocks.Client{}
			sf := NewTokenStrategyFactory(c)

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
			req := getNewEmptyRequest()

			c := &mocks.Client{}
			sf := NewTokenStrategyFactory(c)

			s := sf.Create(authStrategy, testCSRFTokenEndpointURL)

			c.On("GetTokenEndpointResponse", testCSRFTokenEndpointURL, authStrategy).Return(nil, apperrors.NotFound("error"))

			//when
			err := s.AddCSRFToken(req)

			//then
			require.NotNil(t, err)
		})
	})

	t.Run("In case CSRF is disabled", func(t *testing.T) {

		t.Run("Should not modify the original request", func(t *testing.T) {

			// given
			req := getNewEmptyRequest()

			c := &mocks.Client{}
			sf := NewTokenStrategyFactory(c)

			s := sf.Create(nil, "")

			// when
			err := s.AddCSRFToken(req)

			//then
			require.Nil(t, err)
			assert.Empty(t, req.Header)
			assert.Empty(t, req.Cookies())

		})
	})
}

func getNewEmptyRequest() *http.Request {
	return &http.Request{
		Header: make(map[string][]string),
	}
}

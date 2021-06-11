package strategy

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf/mocks"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	authmocks "github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
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
					{Name: cachedCookieName, Value: cachedToken},
				},
			}

			c.On("GetTokenEndpointResponse", testCSRFTokenEndpointURL, authStrategy).Return(cachedItem, nil)

			// when
			err := s.AddCSRFToken(req)

			//then
			require.Nil(t, err)
			cachedCookie, cookieErr := req.Cookie(cachedCookieName)
			require.NoError(t, cookieErr)
			assert.Equal(t, cachedToken, cachedCookie.Value)

		})

		t.Run("Should set CSRF header and merge new Cookies into the request overriding existing cookies", func(t *testing.T) {
			// given
			req := getNewEmptyRequest()
			req.AddCookie(&http.Cookie{Name: cachedCookieName, Value: "oldInvalidCookie"})
			req.AddCookie(&http.Cookie{Name: "custom-user-cookie", Value: "customValue"})

			c := &mocks.Client{}
			sf := NewTokenStrategyFactory(c)

			s := sf.Create(authStrategy, testCSRFTokenEndpointURL)

			cachedItem := &csrf.Response{
				CSRFToken: cachedToken,
				Cookies: []*http.Cookie{
					{Name: cachedCookieName, Value: cachedToken},
				},
			}

			c.On("GetTokenEndpointResponse", testCSRFTokenEndpointURL, authStrategy).Return(cachedItem, nil)

			// when
			err := s.AddCSRFToken(req)

			//then
			require.Nil(t, err)
			assert.Equal(t, cachedToken, req.Header.Get(httpconsts.HeaderCSRFToken))

			cachedCookie, cookieErr := req.Cookie(cachedCookieName)
			require.NoError(t, cookieErr)
			assert.Equal(t, cachedToken, cachedCookie.Value)

			customUserCookie, cookieErr := req.Cookie("custom-user-cookie")
			require.NoError(t, cookieErr)
			assert.Equal(t, "customValue", customUserCookie.Value)

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

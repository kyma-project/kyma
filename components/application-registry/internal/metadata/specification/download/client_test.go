package download

import (
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/csrf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	authHeader = "authorization"
	authValue  = "authValue"
	csrfHeader = "csrfToken"
	csrfValue  = "csrfTokenValue"
)

func TestDownloader_Fetch(t *testing.T) {
	testBody := []byte("testBody")
	t.Run("Should fetch with authorization and token", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, authValue, r.Header.Get(authHeader))
			assert.Equal(t, csrfValue, r.Header.Get(csrfHeader))
			w.Write(testBody)
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		client := &http.Client{}

		downloader := NewClient(client, authFactoryStub{}, csrfFactoryStub{})

		credentials := &authorization.Credentials{}

		bytes, appError := downloader.Fetch(server.URL, credentials)

		require.NoError(t, appError)
		assert.Equal(t, testBody, bytes)
	})

	t.Run("Should fetch without authorization and token when credentials are nil", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEqual(t, authValue, r.Header.Get(authHeader))
			assert.NotEqual(t, csrfValue, r.Header.Get(csrfHeader))
			w.Write(testBody)
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		client := &http.Client{}

		downloader := NewClient(client, authFactoryStub{}, csrfFactoryStub{})

		bytes, appError := downloader.Fetch(server.URL, nil)

		require.NoError(t, appError)
		assert.Equal(t, testBody, bytes)
	})
}

//Authorization stubs
type authFactoryStub struct{}

type authStrategyStub struct{}

func (af authFactoryStub) Create(credentials *authorization.Credentials) authorization.Strategy {
	return authStrategyStub{}
}

func (as authStrategyStub) Invalidate() {}

func (as authStrategyStub) AddAuthorization(r *http.Request, setter authorization.TransportSetter) apperrors.AppError {
	r.Header.Set(authHeader, authValue)
	return nil
}

//CSRF stubs
type csrfFactoryStub struct{}

type csrfStrategyStub struct{}

func (cfs csrfFactoryStub) Create(authorizationStrategy authorization.Strategy, csrfTokenEndpointURL string) csrf.TokenStrategy {
	return csrfStrategyStub{}
}

func (css csrfStrategyStub) AddCSRFToken(apiRequest *http.Request) apperrors.AppError {
	apiRequest.Header.Set(csrfHeader, csrfValue)
	return nil
}

func (css csrfStrategyStub) Invalidate() {}

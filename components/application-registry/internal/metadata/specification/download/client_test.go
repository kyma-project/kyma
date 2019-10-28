package download

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	authHeader = "authorization"
	authValue  = "authValue"
)

func TestDownloader_Fetch(t *testing.T) {
	testBody := []byte("testBody")
	t.Run("Should fetch with authorization", func(t *testing.T) {
		//given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, authValue, r.Header.Get(authHeader))
			w.Write(testBody)
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		client := &http.Client{}

		downloader := NewClient(client, authFactoryStub{})

		credentials := &authorization.Credentials{}
		//when
		bytes, appError := downloader.Fetch(server.URL, credentials, nil)
		//then
		require.NoError(t, appError)
		assert.Equal(t, testBody, bytes)
	})

	t.Run("Should fetch without authorization when credentials are nil", func(t *testing.T) {
		//given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEqual(t, authValue, r.Header.Get(authHeader))
			w.Write(testBody)
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		client := &http.Client{}

		downloader := NewClient(client, authFactoryStub{})
		//when
		bytes, appError := downloader.Fetch(server.URL, nil, nil)
		//then
		require.NoError(t, appError)
		assert.Equal(t, testBody, bytes)
	})

	t.Run("Should add custom headers and query parameters if specified", func(t *testing.T) {
		//given
		headersKey := "headers"
		headersVal := "customHeaders"
		queryKey := "query"
		queryValues := []string{"customParam", "secondParam"}

		customParams := &model.RequestParameters{
			Headers:         &map[string][]string{headersKey: {headersVal}},
			QueryParameters: &map[string][]string{queryKey: queryValues},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			assert.Equal(t, queryValues, query[queryKey])
			assert.Equal(t, headersVal, r.Header.Get(headersKey))
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		client := &http.Client{}

		downloader := NewClient(client, authFactoryStub{})

		//when
		_, appError := downloader.Fetch(server.URL, nil, customParams)

		//then
		require.NoError(t, appError)
	})

	t.Run("Should return error when status code differs from 200", func(t *testing.T) {
		//given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		client := &http.Client{}

		downloader := NewClient(client, authFactoryStub{})
		//when
		_, appError := downloader.Fetch(server.URL, nil, nil)
		//then
		require.Error(t, appError)
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

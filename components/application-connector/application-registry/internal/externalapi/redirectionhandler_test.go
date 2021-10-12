package externalapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectionHandler_Redirect(t *testing.T) {
	t.Run("should redirect request", func(t *testing.T) {
		// given
		redirectionHandler := NewRedirectionHandler("/{application}/v1/metadata/api.yaml", http.StatusMovedPermanently)

		router := mux.NewRouter()
		router.Path("/{application}/v1/metadata").HandlerFunc(redirectionHandler.Redirect)

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		path := "/ec-default/v1/metadata"
		fullUrl := fmt.Sprintf("%s%s", testServer.URL, path)
		req, err := http.NewRequest(http.MethodGet, fullUrl, nil)
		require.NoError(t, err)
		mux.SetURLVars(req, map[string]string{"application": "ec-default"})

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				require.Equal(t, "/ec-default/v1/metadata/api.yaml", req.URL.Path)
				return http.ErrUseLastResponse
			},
		}

		// when
		response, err := client.Get(fullUrl)

		// then
		require.NoError(t, err)
		assert.Equal(t, http.StatusMovedPermanently, response.StatusCode)
	})
}

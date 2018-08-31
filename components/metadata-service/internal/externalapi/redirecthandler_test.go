package externalapi

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gorilla/mux"
	"fmt"
)

func TestRedirectHandler_HandleRequest(t *testing.T) {
	t.Run("should redirect request", func(t *testing.T) {
		// given
		router := mux.NewRouter()
		router.Path("/{remoteEnvironment}/v1/metadata").Handler(NewRedirectHandler("/{remoteEnvironment}/v1/metadataapi.yaml", http.StatusMovedPermanently)).Methods(http.MethodGet)

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		path := "/ec-default/v1/metadata"
		fullUrl := fmt.Sprintf("%s%s", testServer.URL, path)
		req, err := http.NewRequest(http.MethodGet, fullUrl, nil)
		require.NoError(t, err)
		mux.SetURLVars(req, map[string]string{"remoteEnvironment":"ec-default"})

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				require.Equal(t, "/ec-default/v1/metadataapi.yaml", req.URL.Path)
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

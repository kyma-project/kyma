package graphql

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestReadConfig(t *testing.T) {
	t.Run("Should return config from file", func(t *testing.T) {
		//given
		service := NewGraphQLService()

		headers := Headers{}
		headers["faros-xf-user"] = "connector-service"
		headers["faros-xf-groups"] = "kyma-admins"

		expectedConfig := Config{
			URL:     "https://faros.test.graph.ql",
			Headers: headers,
		}

		file, e := os.Open("testdata/config.json")
		require.NoError(t, e)

		//when
		config, e := service.ReadConfig(file)
		require.NoError(t, e)

		//then
		assert.Equal(t, expectedConfig, config)
	})
}

func TestCreateRequest(t *testing.T) {
	t.Run("Should send request with correct query and header", func(t *testing.T) {
		//given
		service := NewGraphQLService()
		header := "testHeader1"
		value := "headerValue1"
		query := `{"query":"{ applications}"}`

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, value, r.Header.Get(header))
			body, e := ioutil.ReadAll(r.Body)
			require.NoError(t, e)
			assert.Equal(t, query, string(body))
			w.WriteHeader(http.StatusOK)
		})

		server := httptest.NewServer(handler)
		defer server.Close()

		config := Config{
			URL:     server.URL,
			Headers: map[string]string{header: value},
		}

		//when
		response, e := service.SendRequest(query, config, time.Duration(1*time.Second))
		require.NoError(t, e)

		//then
		assert.Equal(t, http.StatusOK, response.StatusCode)
	})
}

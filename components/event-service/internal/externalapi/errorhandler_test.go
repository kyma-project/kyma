package externalapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/event-service/internal/httperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorHandler_ServeHTTP(t *testing.T) {
	t.Run("Should always Respond with given error and status code", func(t *testing.T) {

		r := mux.NewRouter()

		r.NotFoundHandler = NewErrorHandler(404, "Requested resource could not be found.")
		ts := httptest.NewServer(r)
		defer ts.Close()

		// when
		res, err := http.Get(ts.URL + "/wrong/path")
		if err != nil {
			assert.Fail(t, "Failure while getting response.")
		}

		responseBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, "Failure while reading response body.")
		}
		defer res.Body.Close()

		var errResponse httperrors.ErrorResponse

		json.Unmarshal(responseBody, &errResponse)

		// then
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, errResponse.Code)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}

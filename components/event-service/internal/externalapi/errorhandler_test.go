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

	v2 "github.com/kyma-project/kyma/components/event-service/internal/externalapi/v2"
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

func Test_filterCEHeaders(t *testing.T) {
	headers := make(map[string][]string)
	request := http.Request{
		Header: headers,
	}
	result := v2.FilterCEHeaders(&request)
	assert.Len(t, result, 0)

	headers["ce-test-header"] = []string{"test-value"}
	result = v2.FilterCEHeaders(&request)
	assert.Len(t, result, 1)
	assert.Equal(t, headers["ce-test-header"][0], "test-value")

	headers["NO-ce-test-header"] = []string{"NO-test-value"}
	result = v2.FilterCEHeaders(&request)
	assert.Len(t, result, 1)
	assert.Equal(t, headers["ce-test-header"][0], "test-value")
}

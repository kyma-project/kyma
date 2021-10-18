package externalapi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/httperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidStateHandler_HandleRequest(t *testing.T) {

	t.Run("Should respond with 500 status code and given error message", func(t *testing.T) {
		// given
		ifh := invalidStateHandler{"initialization error"}

		req, err := http.NewRequest(http.MethodGet, "/re/v1/metadata/services/1234", nil)
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"application": "app", "serviceId": "1234"})
		rr := httptest.NewRecorder()

		// when
		ifh.GetService(rr, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		response := unmarshallResponse(t, rr.Body)
		assert.Contains(t, response.Error, "initialization error")
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})
}

func unmarshallResponse(t *testing.T, body *bytes.Buffer) httperrors.ErrorResponse {
	responseBody, err := ioutil.ReadAll(body)
	if err != nil {
		assert.Fail(t, "Failure while reading response body.")
	}

	var errResponse httperrors.ErrorResponse
	json.Unmarshal(responseBody, &errResponse)

	return errResponse
}

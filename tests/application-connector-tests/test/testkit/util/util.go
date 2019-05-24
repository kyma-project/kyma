package util

import (
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/testkit/services"
)

func RequireStatus(t *testing.T, expectedStatus int, response *http.Response) {
	if expectedStatus != response.StatusCode {
		LogResponse(t, response)
		t.Fatal("Invalid response code")
	}
}

func RequireNoError(t *testing.T, errorResponse *services.ErrorResponse) {
	require.Nil(t, errorResponse)
}

func LogResponse(t *testing.T, resp *http.Response) {
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		t.Logf("failed to dump response, %s", err)
	}

	reqDump, err := httputil.DumpRequest(resp.Request, true)
	if err != nil {
		t.Logf("failed to dump request, %s", err)
	}

	if err == nil {
		t.Logf("\n--------------------------------\n%s\n--------------------------------\n%s\n--------------------------------", reqDump, respDump)
	}
}

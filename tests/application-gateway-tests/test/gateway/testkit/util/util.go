package util

import (
	"net/http"
	"net/http/httputil"
	"testing"
)

func RequireStatus(t *testing.T, expectedStatus int, response *http.Response) {
	if expectedStatus != response.StatusCode {
		LogResponse(t, response)
		t.Fatal("Invalid response code")
	}
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

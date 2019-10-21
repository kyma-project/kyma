package oauth2_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/avast/retry-go"

	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/oauth2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testID        = "test-id"
	testSecret    = "test-secret"
	testScope     = "read"
	testGrantType = "client_credentials"

	testToken                   = "_anDf42CY1-f_lNBvV37TT-aywyasUdeuloKrGX0Ycc.DHtsx0HWIMqF-7s0IHs5kLZJGjSYqlLJgDgJLh8rkGo"
	expiresIn, scope, tokenType = int64(3599), "read", "bearer"

	error400, errorDesc400 = "invalid_request", "the request is missing a required parameter"
	error401, errorDesc401 = "invalid_client", "client authentication failed"
	error500, errorDesc500 = "internal error", "the server encountered an internal error or misconfiguration and was unable to complete your request"

	expectedUnrecoverableErrorMsg = "unrecoverable error, no retries"
)

var (
	testTokenResponse = fmt.Sprintf(`{"access_token":"%s", "expires_in":%d, "scope":"%s", "token_type":"%s"}`, testToken, expiresIn, scope, tokenType)

	commonOpts = []retry.Option{
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
		retry.Delay(time.Millisecond),
	}

	tokenRequest = &oauth2.TokenRequest{
		OAuth2ClientID:     testID,
		OAuth2ClientSecret: testSecret,
		Scope:              testScope,
		GrantType:          testGrantType,
	}

	expectedToken = &oauth2.Token{AccessToken: testToken, ExpiresIn: expiresIn, Scope: scope, TokenType: tokenType}
)

type serverErr struct {
	msg  string
	desc string
}

type testCase struct {
	statusCode            int
	err                   *serverErr
	expectedCalls         int
	shouldFailWith500Once bool
}

func TestFetchOAuth2Token(t *testing.T) {

	t.Run("TestFetchOAuth2Token should", func(t *testing.T) {

		for description, tc := range map[string]testCase{
			"return no error if the request is accepted by Hydra":          {http.StatusOK, nil, 1, false},
			"return an unrecoverable error if the request is malformed":    {http.StatusBadRequest, &serverErr{error400, errorDesc400}, 1, false},
			"return an unrecoverable error if the credentials are invalid": {http.StatusUnauthorized, &serverErr{error401, errorDesc401}, 1, false},
			"return an error and retry if call fails for other reasons":    {http.StatusOK, nil, 2, true},
		} {
			t.Run(description, func(t *testing.T) {

				//given
				handler, counter := getHandlerFunc(t, tc)
				serverURL, _ := runServer(handler)
				mgr := oauth2.NewTokenManager(&http.Client{}, serverURL, commonOpts)

				//when
				retrieved, err := mgr.FetchOAuth2Token(tokenRequest)

				//then
				if tc.statusCode == http.StatusOK {
					require.NoError(t, err)
					assert.Equal(t, expectedToken, retrieved)
				} else {
					require.Error(t, err)
					require.Nil(t, retrieved)
					assert.Contains(t, err.Error(), strconv.Itoa(tc.statusCode))
					assert.Contains(t, err.Error(), tc.err.msg)
					assert.Contains(t, err.Error(), tc.err.desc)

					if tc.unrecoverable() {
						assert.Contains(t, err.Error(), expectedUnrecoverableErrorMsg)
					}
				}

				assert.Equal(t, tc.expectedCalls, *counter)
			})
		}
	})
}

func runServer(h http.HandlerFunc) (*url.URL, error) {
	return url.Parse(httptest.NewServer(h).URL)
}

func getHandlerFunc(t *testing.T, testCase testCase) (func(http.ResponseWriter, *http.Request), *int) {

	assert := assert.New(t)
	var counter int

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		//update number of calls
		counter++

		//verify request
		assert.Equal(http.MethodPost, req.Method)
		assert.Equal("/oauth2/token", req.URL.Path)

		//verify headers
		assert.Equal("application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
		assert.Equal("application/json", req.Header.Get("Accept"))

		//verify auth
		id, secret, ok := req.BasicAuth()
		assert.True(ok)
		assert.Equal(testID, id)
		assert.Equal(testSecret, secret)

		//verify body
		require.NoError(t, req.ParseForm())
		assert.Equal(testScope, req.Form.Get("scope"))
		assert.Equal(testGrantType, req.Form.Get("grant_type"))

		//write response
		if counter == 1 && testCase.shouldFailWith500Once {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(getErrorResponseBody(error500, errorDesc500, http.StatusInternalServerError)))
			return
		}

		w.WriteHeader(testCase.statusCode)
		w.Write([]byte(testCase.getBody()))
	})

	return handler, &counter
}

func (tc *testCase) getBody() string {
	if tc.statusCode == http.StatusOK {
		return testTokenResponse
	}
	return getErrorResponseBody(tc.err.msg, tc.err.desc, tc.statusCode)
}

func (tc *testCase) unrecoverable() bool {
	return tc.statusCode == http.StatusBadRequest || tc.statusCode == http.StatusUnauthorized
}

func getErrorResponseBody(errorMsg, errorDesc string, code int) string {
	return fmt.Sprintf(`{"error":"%s","error_description":"%s","status_code":%d}`, errorMsg, errorDesc, code)
}

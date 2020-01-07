package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/api"
	"github.com/stretchr/testify/assert"
)

const testToken = "ZYqT86bNtVT-QViFpKGsmlnKGpovxVCQ8cMGsQQVU8A.WQC8MchDy-uyW2iIdqW7m26yZwmGAk_I6cR-YO-IiPY"

var unsecuredEndpoint = func() string {
	return runServer(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}()

var securedEndpoint = func() string {
	return runServer(func(w http.ResponseWriter, req *http.Request) {
		token := req.Header.Get("Authorization")

		if token != fmt.Sprintf("Bearer %s", testToken) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}()

var opts = []retry.Option{
	retry.Attempts(3),
	retry.Delay(time.Millisecond),
	retry.DelayType(retry.FixedDelay),
}

var tester = api.NewTester(&http.Client{}, opts)

func TestTestUnsecuredAPI(t *testing.T) {

	t.Run("should return no error if the endpoint is not secured", func(t *testing.T) {

		//when
		err := tester.TestUnsecuredEndpoint(unsecuredEndpoint)

		//then
		assert.NoError(t, err)
	})

	t.Run("should return an error if the endpoint is secured", func(t *testing.T) {

		//when
		err := tester.TestUnsecuredEndpoint(securedEndpoint)

		//then
		assert.Error(t, err)
	})
}

func TestTestSecuredAPI(t *testing.T) {

	for desc, token := range map[string]string{
		//given
		"return an error if no token":   "",
		"return an error if bad token":  "bad_token",
		"return no error if good token": testToken,
	} {
		t.Run(fmt.Sprintf("case: call secured API should %s has been included in the request", desc), func(t *testing.T) {

			//when
			err := tester.TestSecuredEndpoint(securedEndpoint, fmt.Sprintf("Bearer %s", token), "Authorization")

			//then
			if token == testToken {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unexpected response: 401 Unauthorized")
			}
		})
	}
}

func runServer(handler func(w http.ResponseWriter, req *http.Request)) string {
	url, _ := url.Parse(httptest.NewServer(http.HandlerFunc(handler)).URL)
	return url.String()
}

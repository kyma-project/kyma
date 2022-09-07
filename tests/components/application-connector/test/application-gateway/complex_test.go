package application_gateway

import (
	"time"

	"github.com/kyma-project/kyma/tests/components/application-connector/internal/testkit/httpd"
)

func (gs *GatewaySuite) TestComplex() {
	gs.Run("OAuth token renewal", func() {
		http := httpd.NewCli(gs.T())

		url := gatewayURL("complex-cases", "oauth-expired-token-renewal")
		gs.T().Log("Url:", url)

		// Authorize, then call endpoint
		res, _, err := http.Get(url)
		gs.Nil(err, "First request failed")
		gs.Equal(200, res.StatusCode, "First request failed")

		time.Sleep(10 * time.Second) // wait for token to expire

		// Call endpoint, requiring token renewall
		res, _, err = http.Get(url)
		gs.Nil(err, "Second request failed")
		gs.Equal(200, res.StatusCode, "Second request failed")
	})
}

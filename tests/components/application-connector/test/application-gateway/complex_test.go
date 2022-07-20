package application_gateway

import (
	"net/http"
	"time"
)

func (gs *GatewaySuite) TestComplex() {
	gs.Run("OAuth token renewal", func() {
		url := gatewayURL("complex-cases", "oauth-expired-token-renewal")
		gs.T().Log("Url:", url)

		// Authorize, then call endpoint
		res, err := http.Get(url)
		logBody(gs.T(), res.Body)
		gs.Nil(err, "First request failed")
		gs.Equal(200, res.StatusCode, "First request failed")

		time.Sleep(10 * time.Second) // wait for token to expire

		// Call endpoint, requiring token renewall
		res, err = http.Get(url)
		logBody(gs.T(), res.Body)
		gs.Nil(err, "Second request failed")
		gs.Equal(200, res.StatusCode, "Second request failed")
	})
}

package application_gateway

import (
	"net/http"
	"time"
)

func (gs *GatewaySuite) TestManual() {
	gs.Run("OAuth token renewal", func() {
		url := gatewayUrl("manual", "oauth-short")
		gs.T().Log("Url:", url)

		req, err := http.NewRequest(http.MethodDelete, url+"/deauth", nil)
		gs.Nil(err)

		// Authorize, then call endpoint
		// that deauthorizes token used to call it
		res, err := http.DefaultClient.Do(req)
		logBody(gs.T(), res.Body)
		gs.Nil(err, "Deauth request failed")
		gs.Equal(200, res.StatusCode, "Deauth request failed")

		time.Sleep(10 * time.Second) // wait for token to expire

		// Call endpoint requiring authoization
		res, err = http.Get(url + "/ok")
		logBody(gs.T(), res.Body)
		gs.Nil(err, "Ok request failed")
		gs.Equal(200, res.StatusCode, "Ok request failed")
	})
}

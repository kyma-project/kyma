package application_connectivity_validator

func (gs *ValidatorSuite) TestValidator() {
	gs.Run("WIP - positive case", func() {
		gs.True(true)

		////TODO: Use the certificates
		//http := NewHttpCli(gs.T())
		//
		//url := validatorURL("validator-test-app", "httpbin")
		//gs.T().Log("Url:", url)
		//
		//// Authorize, then call endpoint
		//res, _, err := http.Get(url)
		//gs.Nil(err, "Request failed")
		//gs.Equal(200, res.StatusCode, "Request failed")
	})
}

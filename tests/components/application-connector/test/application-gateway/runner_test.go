package application_gateway

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var applications = []string{"positive-authorisation", "negative-authorisation", "path-related-error-handling", "missing-resources-error-handling", "proxy-cases"}

func (gs *GatewaySuite) TestGetRequest() {

	for _, app := range applications {
		app, err := gs.cli.ApplicationconnectorV1alpha1().Applications().Get(context.Background(), app, v1.GetOptions{})
		gs.Nil(err)

		gs.Run(app.Spec.Description, func() {
			for _, service := range app.Spec.Services {
				gs.Run(service.Description, func() {
					http := NewHttpCli(gs.T())

					for _, entry := range service.Entries {
						if entry.Type != "API" {
							gs.T().Log("Skipping event entry")
							continue
						}

						expectedCode, err := getExpectedHTTPCode(service)
						if err != nil {
							gs.T().Log("Error during getting the error code from description -> applicationCRD")
							gs.T().Fail()
						}

						res, _, err := http.Get(entry.CentralGatewayUrl)
						gs.Nil(err, "Request failed")
						gs.Equal(expectedCode, res.StatusCode, "Incorrect response code")
					}
				})
			}
		})
	}
}

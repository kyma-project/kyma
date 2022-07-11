package application_gateway_test

import (
	"context"
	"io/ioutil"
	"net/http"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (gs *GatewaySuite) TestSimpleCases() {
	app, err := gs.cli.ApplicationconnectorV1alpha1().Applications().Get(context.Background(), "test-app", v1.GetOptions{})

	gs.Nil(err)

	for _, service := range app.Spec.Services {
		gs.Run(service.DisplayName, func() {
			for _, entry := range service.Entries {
				if entry.Type != "API" {
					gs.T().Log("Skipping event entry.")
					continue
				}
				gs.T().Log("Calling", entry.CentralGatewayUrl)
				res, err := http.Get(entry.CentralGatewayUrl)
				gs.Nil(err)
				if err == nil {
					body, err := ioutil.ReadAll(res.Body)
					if err == nil && len(body) > 0 {
						gs.T().Log("Response", string(body))
					}
					gs.Equal(200, res.StatusCode)
				}
			}
		})
	}
}

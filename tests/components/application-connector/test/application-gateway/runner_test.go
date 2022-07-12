package application_gateway

import (
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var applications = []string{"positive-authorisation", "path-related-error-handling", "kubernetes-resources-error-handling"}

func (gs *GatewaySuite) TestCases() {

	for _, app := range applications {
		app, err := gs.cli.ApplicationconnectorV1alpha1().Applications().Get(context.Background(), app, v1.GetOptions{})
		gs.Nil(err)

		for _, service := range app.Spec.Services {
			gs.Run(service.Description, func() {
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

					actualCode := executeGetRequest(gs.T(), entry)

					gs.Equal(expectedCode, actualCode)
				}
			})
		}
	}
}

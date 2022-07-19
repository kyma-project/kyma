package application_gateway

import (
	"context"
	"io/ioutil"
	"strconv"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (gs *GatewaySuite) TestResponseBody() {
	app, err := gs.cli.ApplicationconnectorV1alpha1().Applications().Get(context.Background(), "proxy-cases", v1.GetOptions{})
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

				actualCode, res := executeGetRequest(gs.T(), entry)

				body, err := ioutil.ReadAll(res.Body)
				gs.Nil(err)
				codeStr := strconv.Itoa(expectedCode)
				gs.Equal(codeStr, string(body))

				gs.Equal(expectedCode, actualCode)
			}
		})
	}
}

package application_gateway

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/components/application-connector/internal/testkit/httpd"
)

var applications = []string{"positive-authorisation", "negative-authorisation", "path-related-error-handling", "missing-resources-error-handling", "proxy-cases", "proxy-errors", "redirects", "code-rewriting"}

func (gs *GatewaySuite) TestGetRequest() {

	for _, app := range applications {
		app, err := gs.cli.ApplicationconnectorV1alpha1().Applications().Get(context.Background(), app, v1.GetOptions{})
		gs.Nil(err)

		gs.Run(app.Spec.Description, func() {
			for _, service := range app.Spec.Services {
				gs.Run(service.Description, func() {
					http := httpd.NewCli(gs.T())

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

func (gs *GatewaySuite) TestResponseBody() {
	app, err := gs.cli.ApplicationconnectorV1alpha1().Applications().Get(context.Background(), "proxy-cases", v1.GetOptions{})
	gs.Nil(err)
	for _, service := range app.Spec.Services {
		gs.Run(service.Description, func() {
			http := httpd.NewCli(gs.T())

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

				_, body, err := http.Get(entry.CentralGatewayUrl)
				gs.Nil(err, "Request failed")

				codeStr := strconv.Itoa(expectedCode)

				gs.Equal(codeStr, string(body), "Incorrect body")
			}
		})
	}
}

func (gs *GatewaySuite) TestBodyPerMethod() {
	app, err := gs.cli.ApplicationconnectorV1alpha1().Applications().Get(context.Background(), "methods-with-body", v1.GetOptions{})
	gs.Nil(err)
	for _, service := range app.Spec.Services {
		gs.Run(service.Description, func() {
			httpCli := httpd.NewCli(gs.T())

			for _, entry := range service.Entries {
				if entry.Type != "API" {
					gs.T().Log("Skipping event entry")
					continue
				}

				method := service.Description
				bodyBuf := strings.NewReader(service.Description)

				req, err := http.NewRequest(method, entry.CentralGatewayUrl, bodyBuf)
				gs.Nil(err, "Preparing request failed")

				_, body, err := httpCli.Do(req)
				gs.Nil(err, "Request failed")

				res, err := unmarshalBody(body)
				gs.Nil(err, "Response body wasn't correctly forwarded")

				gs.Equal(service.Description, string(res.Body), "Request body doesn't match")
				gs.Equal(service.Description, res.Method, "Request method doesn't match")
			}
		})
	}
}

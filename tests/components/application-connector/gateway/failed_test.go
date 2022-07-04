package main_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (gs *GatewaySuite) TestFailedCases() {
	app, err := gs.cli.ApplicationconnectorV1alpha1().
		Applications().
		Get(context.Background(), "failed-auth-cases", v1.GetOptions{})

	gs.Nil(err)

	for _, service := range app.Spec.Services {
		code, err := codeFromService(service)
		gs.Nil(err)
		gs.Run(service.DisplayName, func() {
			for _, entry := range service.Entries {
				if entry.Type != "API" {
					gs.T().Log("Skipping event entry")
					continue
				}
				gs.T().Log("Calling", entry.CentralGatewayUrl)
				res, err := http.Get(entry.CentralGatewayUrl)
				gs.Nil(err)
				body, err := ioutil.ReadAll(res.Body)
				if err == nil && len(body) > 0 {
					gs.T().Log("Response", string(body))
				}
				gs.Equal(code, res.StatusCode)
			}
		})
	}
}

// codeFromService extracts expected response code from
// service's description. Default to 403
func codeFromService(service v1alpha1.Service) (code int, err error) {
	code = 403
	re := regexp.MustCompile(`\d+$`) // NOTE: Is this a terrible idea?
	if codeStr := re.FindString(service.Description); len(codeStr) > 0 {
		code, err = strconv.Atoi(codeStr)
		if err != nil {
			return
		}
	}

	return
}

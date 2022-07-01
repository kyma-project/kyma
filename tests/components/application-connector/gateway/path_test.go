package main_test

import (
	"context"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func callToGateway(t *testing.T, entry v1alpha1.Entry) int {
	t.Log("Calling", entry.CentralGatewayUrl)
	res, err := http.Get(entry.CentralGatewayUrl)
	assert.Nil(t, err)
	body, err := ioutil.ReadAll(res.Body)
	if err == nil && len(body) > 0 {
		t.Log("Response", string(body))
	}
	return res.StatusCode
}

func (gs *GatewaySuite) TestPath() {
	app, err := gs.cli.ApplicationconnectorV1alpha1().Applications().Get(context.Background(), "bad-path", v1.GetOptions{})

	gs.Nil(err)

	for _, service := range app.Spec.Services {
		gs.Run(service.DisplayName, func() {
			for _, entry := range service.Entries {
				if entry.Type != "API" {
					gs.T().Log("Skipping event entry")
					continue
				}
				var code int

				switch service.Name {
				case "missingSrvApp":
					code = callToGateway(gs.T(), entry)
					gs.Equal(500, code)
				case "missingSrv":
					code = callToGateway(gs.T(), entry)
					gs.Equal(500, code)
				case "badTargetURL":
					code = callToGateway(gs.T(), entry)
					gs.Equal(404, code)
				}
			}
		})
	}
}

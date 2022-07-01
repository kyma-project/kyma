package main_test

import (
	"context"
	"github.com/kyma-project/kyma/tests/components/application-connector/gateway/helper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
					code = helper.CallToGateway(gs.T(), entry)
					gs.Equal(500, code)
				case "missingSrv":
					code = helper.CallToGateway(gs.T(), entry)
					gs.Equal(500, code)
				case "badTargetURL":
					code = helper.CallToGateway(gs.T(), entry)
					gs.Equal(404, code)
				}
			}
		})
	}
}

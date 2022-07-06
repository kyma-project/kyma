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
				var actualCode, expectedCode int

				switch service.Name {
				case "missingSrvApp":
					actualCode, expectedCode = helper.GetCodes(gs.T(), entry, service)
					gs.Equal(expectedCode, actualCode)
				case "missingSrv":
					actualCode, expectedCode = helper.GetCodes(gs.T(), entry, service)
					gs.Equal(expectedCode, actualCode)
				case "badTargetURL":
					actualCode, expectedCode = helper.GetCodes(gs.T(), entry, service)
					gs.Equal(expectedCode, actualCode)
				}
			}
		})
	}
}

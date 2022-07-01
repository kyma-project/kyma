package main_test

import (
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const applicationName = "test-app"

func (gs *GatewaySuite) TestYaml() {
	app, err := gs.cli.ApplicationconnectorV1alpha1().Applications().Get(context.Background(), applicationName, v1.GetOptions{})
	namespace := "test"
	expectedTargetURL := "http://" + app.Name + "." + namespace + ".svc.cluster.local:8080"
	gs.Nil(err)

	for _, service := range app.Spec.Services {
		gs.Run(service.DisplayName, func() {
			for _, entry := range service.Entries {
				if entry.Type == "API" {
					gs.Equal(true, strings.Contains(entry.CentralGatewayUrl, service.Name), "Service name not provided in path")
					gs.Equal(true, strings.Contains(entry.CentralGatewayUrl, app.Name), "Application name not provided in path")
					gs.Equal(true, strings.Contains(entry.TargetUrl, namespace), "Namespace not provided in target path")
					gs.Equal(true, strings.Contains(entry.TargetUrl, expectedTargetURL), "Bad targetURL path")
				}
			}
		})
	}
}

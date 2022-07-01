package main_test

import (
	"context"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const applicationName = "test-app"

func (gs *GatewaySuite) TestApplicationName() {
	app, err := gs.cli.ApplicationconnectorV1alpha1().Applications().Get(context.Background(), applicationName, v1.GetOptions{})
	namespace := "test"
	gs.Nil(err)

	for _, service := range app.Spec.Services {
		gs.Run(service.DisplayName, func() {
			for _, entry := range service.Entries {
				if entry.Type == "API" {
					fmt.Println("****************")
					fmt.Println(entry.CentralGatewayUrl)
					fmt.Println(service.Name)
					fmt.Println(app.Name)
					gs.Equal(true, strings.Contains(entry.CentralGatewayUrl, service.Name), "Service name not provided in path")
					gs.Equal(true, strings.Contains(entry.CentralGatewayUrl, app.Name), "Application name not provided in path")
					fmt.Println("****************")

				}
			}
		})
	}

	for _, service := range app.Spec.Services {
		gs.Run(service.DisplayName, func() {
			for _, entry := range service.Entries {
				if entry.Type == "API" {
					fmt.Println("&&&&&&&&&&")
					fmt.Println(entry.CentralGatewayUrl)
					fmt.Println(service.Name)

					gs.Equal(true, strings.Contains(entry.CentralGatewayUrl, service.Name), "Service name not provided in path")
					fmt.Println("&&&&&&&&&&")
				}
			}
		})
	}

	for _, service := range app.Spec.Services {

		expectedTargetURL := "http://" + app.Name + "." + namespace + ".svc.cluster.local:8080"
		gs.Run(service.DisplayName, func() {
			for _, entry := range service.Entries {
				if entry.Type == "API" {
					fmt.Println("/////////////")
					fmt.Println(entry.TargetUrl)
					fmt.Println(service.Name)
					fmt.Println(namespace)
					fmt.Println(expectedTargetURL)

					gs.Equal(true, strings.Contains(entry.TargetUrl, service.Name), "Service name not provided in target path")
					gs.Equal(true, strings.Contains(entry.TargetUrl, namespace), "Namespace not provided in target path")
					gs.Equal(true, strings.Contains(entry.TargetUrl, expectedTargetURL), "Bad targetURL path")
					fmt.Println("///////////////")
				}
			}
		})
	}
}

package main

import (
	"flag"
	"fmt"
)

type options struct {
	appName                string
	domainName             string
	namespace              string
	tillerUrl              string
	gatewayImage           string
	eventServiceImage      string
	eventServiceTestsImage string
}

func parseArgs() *options {
	appName := flag.String("appName", "remote-environment-controller", "Name used in controller registration")
	domainName := flag.String("domainName", "kyma.local", "Domain name of the cluster")
	namespace := flag.String("namespace", "kyma-integration", "Namespace in which the RE chart will be installed")
	tillerUrl := flag.String("tillerUrl", "tiller-deploy.kube-system.svc.cluster.local:44134", "Tiller release server url")

	gatewayImage := flag.String("gatewayImage", "", "The image of the Gateway to use")
	eventServiceImage := flag.String("eventServiceImage", "", "The image of the Event Service to use")
	eventServiceTestsImage := flag.String("eventServiceTestsImage", "", "The image of the Event Service Tests to use")

	flag.Parse()

	return &options{
		appName:                *appName,
		domainName:             *domainName,
		namespace:              *namespace,
		tillerUrl:              *tillerUrl,
		gatewayImage:           *gatewayImage,
		eventServiceImage:      *eventServiceImage,
		eventServiceTestsImage: *eventServiceTestsImage,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--appName=%s --domainName=%s --namespace=%s --tillerUrl=%s "+
		"--gatewayImage=%s --eventServiceImage=%s --eventServiceTestsImage=%s",
		o.appName, o.domainName, o.namespace, o.tillerUrl,
		o.gatewayImage, o.eventServiceImage, o.eventServiceTestsImage)
}

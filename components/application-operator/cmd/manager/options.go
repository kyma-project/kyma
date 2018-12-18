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
	syncPeriod             int
	installationTimeout    int64
	proxyServiceImage      string
	eventServiceImage      string
	eventServiceTestsImage string
}

func parseArgs() *options {
	appName := flag.String("appName", "application-operator", "Name used in controller registration")
	domainName := flag.String("domainName", "kyma.local", "Domain name of the cluster")
	namespace := flag.String("namespace", "kyma-integration", "Namespace in which the RE chart will be installed")
	tillerUrl := flag.String("tillerUrl", "tiller-deploy.kube-system.svc.cluster.local:44134", "Tiller release server url")
	syncPeriod := flag.Int("syncPeriod", 30, "Time period between resyncing existing resources")
	installationTimeout := flag.Int64("installationTimeout", 240, "Time after the release installation will time out")

	proxyServiceImage := flag.String("proxyServiceImage", "", "The image of the Proxy Service to use")
	eventServiceImage := flag.String("eventServiceImage", "", "The image of the Event Service to use")
	eventServiceTestsImage := flag.String("eventServiceTestsImage", "", "The image of the Event Service Tests to use")

	flag.Parse()

	return &options{
		appName:                *appName,
		domainName:             *domainName,
		namespace:              *namespace,
		tillerUrl:              *tillerUrl,
		syncPeriod:             *syncPeriod,
		installationTimeout:    *installationTimeout,
		proxyServiceImage:      *proxyServiceImage,
		eventServiceImage:      *eventServiceImage,
		eventServiceTestsImage: *eventServiceTestsImage,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--appName=%s --domainName=%s --namespace=%s --tillerUrl=%s --syncPeriod=%d --installationTimeout=%d "+
		"--proxyServiceImage=%s --eventServiceImage=%s --eventServiceTestsImage=%s",
		o.appName, o.domainName, o.namespace, o.tillerUrl, o.syncPeriod, o.installationTimeout,
		o.proxyServiceImage, o.eventServiceImage, o.eventServiceTestsImage)
}

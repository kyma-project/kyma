package main

import (
	"flag"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultNamespace = "default"
)

type options struct {
	appName                     string
	namespace                   string
	clusterCertificatesSecret   types.NamespacedName
	caCertificatesSecret        types.NamespacedName
	controllerSyncPeriod        int
	minimalConnectionSyncPeriod int
}

func parseArgs() *options {
	appName := flag.String("appName", "connectivity-certs-controller", "Name used in controller registration")
	namespace := flag.String("namespace", "kyma-integration", "Namespace in which secrets are created")
	clusterCertificatesSecret := flag.String("clusterCertificatesSecret", "kyma-integration/cluster-client-certificates", "Secret namespace/name where cluster client certificate and key are kept")
	caCertificatesSecret := flag.String("caCertificatesSecret", "istio-system/ca-certificates", "Secret namespace/name where CA certificate is kept")
	controllerSyncPeriod := flag.Int("controllerSyncPeriod", 300, "Time period between resyncing existing resources")
	minimalConnectionSyncPeriod := flag.Int("minimalConnectionSyncPeriod", 300, "Minimal time between trying to synchronize with Central Connector Service")

	flag.Parse()

	return &options{
		appName:                     *appName,
		namespace:                   *namespace,
		clusterCertificatesSecret:   parseNamespacedName(*clusterCertificatesSecret),
		caCertificatesSecret:        parseNamespacedName(*caCertificatesSecret),
		controllerSyncPeriod:        *controllerSyncPeriod,
		minimalConnectionSyncPeriod: *minimalConnectionSyncPeriod,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--appName=%s --namespace=%s"+
		"--clusterCertificatesSecret=%s --caCertificatesSecret=%s "+
		"--controllerSyncPeriod=%d --minimalConnectionSyncPeriod=%d",
		o.appName, o.namespace,
		o.clusterCertificatesSecret, o.caCertificatesSecret,
		o.controllerSyncPeriod, o.minimalConnectionSyncPeriod)
}

func parseNamespacedName(value string) types.NamespacedName {
	parts := strings.Split(value, string(types.Separator))

	if len(parts) == 1 {
		return types.NamespacedName{
			Name:      parts[0],
			Namespace: defaultNamespace,
		}
	}

	return types.NamespacedName{
		Namespace: get(parts, 0),
		Name:      get(parts, 1),
	}
}

func get(array []string, index int) string {
	if len(array) > index {
		return array[index]
	}
	return ""
}

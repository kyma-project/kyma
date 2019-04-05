package main

import (
	"flag"
	"fmt"
)

type options struct {
	appName                   string
	namespace                 string
	clusterCertificatesSecret string
	caCertificatesSecret      string
	syncPeriod                int
}

func parseArgs() *options {
	appName := flag.String("appName", "connectivity-certs-controller", "Name used in controller registration")
	namespace := flag.String("namespace", "kyma-integration", "Namespace in which secrets are created")
	clusterCertificatesSecret := flag.String("clusterCertificatesSecret", "cluster-client-certificates", "Secret name where cluster client certificate and key are kept")
	caCertificatesSecret := flag.String("caCertificatesSecret", "ca-certificates", "Secret name where CA certificate is kept")
	syncPeriod := flag.Int("syncPeriod", 300, "Time period between resyncing existing resources")

	flag.Parse()

	return &options{
		appName:                   *appName,
		namespace:                 *namespace,
		clusterCertificatesSecret: *clusterCertificatesSecret,
		caCertificatesSecret:      *caCertificatesSecret,
		syncPeriod:                *syncPeriod,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--appName=%s --namespace=%s"+
		"--clusterCertificatesSecret=%s --caCertificatesSecret=%s --syncPeriod=%d",
		o.appName, o.namespace,
		o.clusterCertificatesSecret, o.caCertificatesSecret, o.syncPeriod)
}

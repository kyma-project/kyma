package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort       int
	proxyPort             int
	minioURL              string
	namespace             string
	requestTimeout        int
	requestLogging        bool
	specRequestTimeout    int
	detailedErrorResponse bool
	centralGatewayUrl     string
	insecureAssetDownload bool
	insecureSpecDownload  bool
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	proxyPort := flag.Int("proxyPort", 8080, "Proxy port.")
	namespace := flag.String("namespace", "kyma-integration", "Namespace used by Application Registry")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	specRequestTimeout := flag.Int("specRequestTimeout", 20, "Timeout for Spec Service.")
	detailedErrorResponse := flag.Bool("detailedErrorResponse", false, "Flag for showing full internal error response messages.")
	centralGatewayUrl := flag.String("centralGatewayUrl", "http://central-application-gateway.kyma-system:8080", "Central Application Gateway URL.")
	insecureAssetDownload := flag.Bool("insecureAssetDownload", false, "Flag for skipping certificate verification for asset download.")
	insecureSpecDownload := flag.Bool("insecureSpecDownload", false, "Flag for skipping certificate verification for API specification download.")

	flag.Parse()

	return &options{
		externalAPIPort:       *externalAPIPort,
		proxyPort:             *proxyPort,
		namespace:             *namespace,
		requestTimeout:        *requestTimeout,
		requestLogging:        *requestLogging,
		specRequestTimeout:    *specRequestTimeout,
		detailedErrorResponse: *detailedErrorResponse,
		centralGatewayUrl:     *centralGatewayUrl,
		insecureAssetDownload: *insecureAssetDownload,
		insecureSpecDownload:  *insecureSpecDownload,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --proxyPort=%d --centralGatewayUrl=%s "+
		"--namespace=%s --requestTimeout=%d  --requestLogging=%t --specRequestTimeout=%d "+
		"--detailedErrorResponse=%t --insecureAssetDownload=%t --insecureSpecDownload=%t",
		o.externalAPIPort, o.proxyPort, o.centralGatewayUrl,
		o.namespace, o.requestTimeout, o.requestLogging, o.specRequestTimeout, o.detailedErrorResponse, o.insecureAssetDownload, o.insecureSpecDownload)
}

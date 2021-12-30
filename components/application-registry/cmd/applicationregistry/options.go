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
	detailedErrorResponse bool
	uploadServiceURL      string
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
	detailedErrorResponse := flag.Bool("detailedErrorResponse", false, "Flag for showing full internal error response messages.")
	uploadServiceURL := flag.String("uploadServiceURL", "http://rafter-upload-service.kyma-system.svc.cluster.local:80", "Upload Service URL.")
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
		detailedErrorResponse: *detailedErrorResponse,
		uploadServiceURL:      *uploadServiceURL,
		centralGatewayUrl:     *centralGatewayUrl,
		insecureAssetDownload: *insecureAssetDownload,
		insecureSpecDownload:  *insecureSpecDownload,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --proxyPort=%d --uploadServiceURL=%s --centralGatewayUrl=%s "+
		"--namespace=%s --requestTimeout=%d  --requestLogging=%t"+
		"--detailedErrorResponse=%t --insecureAssetDownload=%t --insecureSpecDownload=%t",
		o.externalAPIPort, o.proxyPort, o.uploadServiceURL, o.centralGatewayUrl,
		o.namespace, o.requestTimeout, o.requestLogging, o.detailedErrorResponse, o.insecureAssetDownload, o.insecureSpecDownload)
}

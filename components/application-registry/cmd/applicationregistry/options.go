package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort          int
	proxyPort                int
	minioURL                 string
	namespace                string
	requestTimeout           int
	requestLogging           bool
	specRequestTimeout       int
	assetstoreRequestTimeout int
	detailedErrorResponse    bool
	uploadServiceURL         string
	insecureAssetDownload    bool
	insecureSpecDownload     bool
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	proxyPort := flag.Int("proxyPort", 8080, "Proxy port.")
	namespace := flag.String("namespace", "kyma-integration", "Namespace used by Application Registry")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	specRequestTimeout := flag.Int("specRequestTimeout", 5, "Timeout for Spec Service.")
	assetstoreRequestTimeout := flag.Int("assetstoreRequestTimeout", 5, "Timeout for Asset Store Service.")
	detailedErrorResponse := flag.Bool("detailedErrorResponse", false, "Flag for showing full internal error response messages.")
	uploadServiceURL := flag.String("uploadServiceURL", "http://assetstore-asset-upload-service.kyma-system.svc.cluster.local:3000", "Upload Service URL.")
	insecureAssetDownload := flag.Bool("insecureAssetDownload", false, "Flag for skipping certificate verification for asset download. ")
	insecureSpecDownload := flag.Bool("insecureSpecDownload", false, "Flag for skipping certificate verification for API specification download. ")

	flag.Parse()

	return &options{
		externalAPIPort:          *externalAPIPort,
		proxyPort:                *proxyPort,
		namespace:                *namespace,
		requestTimeout:           *requestTimeout,
		requestLogging:           *requestLogging,
		specRequestTimeout:       *specRequestTimeout,
		assetstoreRequestTimeout: *assetstoreRequestTimeout,
		detailedErrorResponse:    *detailedErrorResponse,
		uploadServiceURL:         *uploadServiceURL,
		insecureAssetDownload:    *insecureAssetDownload,
		insecureSpecDownload:     *insecureSpecDownload,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --proxyPort=%d --uploadServiceURL=%s"+
		"--namespace=%s --requestTimeout=%d  --requestLogging=%t --specRequestTimeout=%d"+
		"--assetstoreRequestTimeout=%d --detailedErrorResponse=%t --insecureAssetDownload=%t --insecureSpecDownload=%t",
		o.externalAPIPort, o.proxyPort, o.uploadServiceURL,
		o.namespace, o.requestTimeout, o.requestLogging, o.specRequestTimeout, o.assetstoreRequestTimeout, o.detailedErrorResponse, o.insecureAssetDownload, o.insecureSpecDownload)
}

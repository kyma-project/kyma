package externalapi

import (
	"fmt"

	"github.com/kyma-project/kyma/components/connector-service/internal/httpcontext"
)

const (
	MetadataURLFormat = "https://%s/%s/v1/metadata/services"
	EventsURLFormat   = "https://%s/%s/v1/events"
)

type APIUrlsGenerator interface {
	Generate(reader httpcontext.ContextReader) interface{}
}

type applicationUrlsStrategy struct {
	appRegistryHost string
	eventsHost      string
	infoURL         string
	certURL         string
}

func (appInfoUrls applicationUrlsStrategy) Generate(reader httpcontext.ContextReader) interface{} {
	return applicationApi{
		MetadataURL:     fmt.Sprintf(MetadataURLFormat, appInfoUrls.appRegistryHost, reader.GetApplication()),
		EventsURL:       fmt.Sprintf(EventsURLFormat, appInfoUrls.eventsHost, reader.GetApplication()),
		InfoURL:         appInfoUrls.infoURL,
		CertificatesURL: appInfoUrls.certURL,
	}
}

func NewApplicationApiUrlsStrategy(appRegistryHost, eventsHost, inforURL, certURL string) APIUrlsGenerator {
	return &applicationUrlsStrategy{
		appRegistryHost: appRegistryHost,
		eventsHost:      eventsHost,
		infoURL:         inforURL,
		certURL:         certURL,
	}
}

type runtimeUrlsStrategy struct {
	infoURL string
	certURL string
}

func (runtimeInfoUrls runtimeUrlsStrategy) Generate(reader httpcontext.ContextReader) interface{} {
	return runtimeApi{
		InfoURL:         runtimeInfoUrls.infoURL,
		CertificatesURL: runtimeInfoUrls.certURL,
	}
}

func NewRuntimeApiUrlsStrategy(infoURL, certURL string) APIUrlsGenerator {
	return &runtimeUrlsStrategy{
		infoURL: infoURL,
		certURL: certURL,
	}
}

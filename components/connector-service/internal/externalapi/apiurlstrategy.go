package externalapi

import (
	"fmt"

	"github.com/kyma-project/kyma/components/connector-service/internal/httpcontext"
)

const (
	MetadataURLFormat          = "https://%s/%s/v1/metadata/services"
	EventsURLFormat            = "https://%s/%s/v1/events"
	AppManagementInfoURLFormat = "https://%s/v1/applications/management/info"
	AppCertificatesURLFormat   = "https://%s/v1/applications/certificates"

	RuntimeManagementInfoURLFormat = "https://%s/v1/runtimes/management/info"
	RuntimeCertificatesURLFormat   = "https://%s/v1/runtimes/certificates"
)

type APIUrlsGenerator interface {
	Generate(reader httpcontext.ConnectorClientReader) interface{}
}

type applicationUrlsStrategy struct {
	appRegistryHost string
	eventsHost      string
	infoURL         string
	certURL         string
}

func (appInfoUrls applicationUrlsStrategy) Generate(reader httpcontext.ConnectorClientReader) interface{} {
	return applicationApi{
		MetadataURL:     fmt.Sprintf(MetadataURLFormat, appInfoUrls.appRegistryHost, reader.GetApplication()),
		EventsURL:       fmt.Sprintf(EventsURLFormat, appInfoUrls.eventsHost, reader.GetApplication()),
		InfoURL:         appInfoUrls.infoURL,
		CertificatesURL: appInfoUrls.certURL,
	}
}

func NewApplicationApiUrlsStrategy(appRegistryHost, eventsHost, infoURL, host string) APIUrlsGenerator {
	if infoURL == "" {
		infoURL = fmt.Sprintf(AppManagementInfoURLFormat, host)
	}

	return &applicationUrlsStrategy{
		appRegistryHost: appRegistryHost,
		eventsHost:      eventsHost,
		infoURL:         infoURL,
		certURL:         fmt.Sprintf(AppCertificatesURLFormat, host),
	}
}

type runtimeUrlsStrategy struct {
	infoURL string
	certURL string
}

func (runtimeInfoUrls runtimeUrlsStrategy) Generate(reader httpcontext.ConnectorClientReader) interface{} {
	return runtimeApi{
		InfoURL:         runtimeInfoUrls.infoURL,
		CertificatesURL: runtimeInfoUrls.certURL,
	}
}

func NewRuntimeApiUrlsStrategy(host string) APIUrlsGenerator {
	return &runtimeUrlsStrategy{
		infoURL: fmt.Sprintf(RuntimeManagementInfoURLFormat, host),
		certURL: fmt.Sprintf(RuntimeCertificatesURLFormat, host),
	}
}

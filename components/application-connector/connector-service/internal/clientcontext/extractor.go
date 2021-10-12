package clientcontext

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

const (
	MetadataURLFormat   = "%s/%s/v1/metadata/services"
	EventsURLFormat     = "%s/%s/v1/events"
	EventsInfoURLFormat = "%s/%s/v1/events/subscribed"

	RuntimeDefaultCommonName = "*Runtime*"
)

type ConnectorClientExtractor func(ctx context.Context) (ClientCertContextService, apperrors.AppError)

type ApplicationContextExtractor func(ctx context.Context) (ApplicationContext, apperrors.AppError)

type ContextExtractor struct {
	subjectDefaults certificates.CSRSubject
}

func NewContextExtractor(subjectDefaults certificates.CSRSubject) *ContextExtractor {
	return &ContextExtractor{
		subjectDefaults: subjectDefaults,
	}
}

func (ext *ContextExtractor) CreateApplicationClientContextService(ctx context.Context) (ClientCertContextService, apperrors.AppError) {
	appCtx, err := ExtractApplicationContext(ctx)
	if err != nil {
		return nil, err
	}

	subject := ext.prepareSubject(appCtx.Tenant, appCtx.Group, appCtx.Application)

	apiHosts, ok := ctx.Value(ApiURLsKey).(ApiURLs)
	if !ok {
		return newClientCertificateContext(appCtx, subject), nil
	}

	extendedCtx := ExtendedApplicationContext{
		ApplicationContext: appCtx,
		RuntimeURLs:        prepareRuntimeURLs(appCtx, apiHosts),
	}

	return newClientCertificateContext(extendedCtx, subject), nil
}

func prepareRuntimeURLs(appCtx ApplicationContext, apiHosts ApiURLs) RuntimeURLs {
	metadataURL := ""
	eventsURL := ""
	eventsInfoURL := ""

	if apiHosts.MetadataBaseURL != "" {
		metadataURL = fmt.Sprintf(MetadataURLFormat, apiHosts.MetadataBaseURL, appCtx.GetApplication())
	}

	if apiHosts.EventsBaseURL != "" {
		eventsURL = fmt.Sprintf(EventsURLFormat, apiHosts.EventsBaseURL, appCtx.GetApplication())
		eventsInfoURL = fmt.Sprintf(EventsInfoURLFormat, apiHosts.EventsBaseURL, appCtx.GetApplication())
	}

	return RuntimeURLs{
		MetadataURL:   metadataURL,
		EventsURL:     eventsURL,
		EventsInfoURL: eventsInfoURL,
	}
}

func (ext *ContextExtractor) prepareSubject(org, orgUnit, commonName string) certificates.CSRSubject {
	organization := org
	organizationalUnit := orgUnit

	if isEmpty(organization) {
		organization = ext.subjectDefaults.Organization
	}

	if isEmpty(organizationalUnit) {
		organizationalUnit = ext.subjectDefaults.OrganizationalUnit
	}

	return certificates.CSRSubject{
		Organization:       organization,
		OrganizationalUnit: organizationalUnit,
		CommonName:         commonName,
		Country:            ext.subjectDefaults.Country,
		Locality:           ext.subjectDefaults.Locality,
		Province:           ext.subjectDefaults.Province,
	}
}

func (ext *ContextExtractor) CreateClusterClientContextService(ctx context.Context) (ClientCertContextService, apperrors.AppError) {
	clusterCtx, ok := ctx.Value(ClusterContextKey).(ClusterContext)
	if !ok {
		return nil, apperrors.Internal("Failed to extract ClusterContext from request")
	}

	subject := ext.prepareSubject(clusterCtx.Tenant, clusterCtx.Group, RuntimeDefaultCommonName)

	return newClientCertificateContext(clusterCtx, subject), nil
}

func ExtractApplicationContext(ctx context.Context) (ApplicationContext, apperrors.AppError) {
	appCtx, ok := ctx.Value(ApplicationContextKey).(ApplicationContext)
	if !ok {
		return ApplicationContext{}, apperrors.Internal("Failed to extract ApplicationContext from request")
	}
	return appCtx, nil
}

func isEmpty(str string) bool {
	return str == ""
}

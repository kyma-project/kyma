package clientcontext

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

const (
	MetadataURLFormat = "https://%s/%s/v1/metadata/services"
	EventsURLFormat   = "https://%s/%s/v1/events"

	RuntimeDefaultCommonName = "Runtime"
)

type ConnectorClientExtractor func(ctx context.Context) (ClientContextService, apperrors.AppError)

type ApplicationContextExtractor func(ctx context.Context) (ApplicationContext, apperrors.AppError)

type ContextExtractor struct {
	subjectDefaults certificates.CSRSubject
}

func NewContextExtractor(subjectDefaults certificates.CSRSubject) *ContextExtractor {
	return &ContextExtractor{
		subjectDefaults: subjectDefaults,
	}
}

func (ext *ContextExtractor) CreateApplicationClientContextService(ctx context.Context) (ClientContextService, apperrors.AppError) {
	appCtx, err := ExtractApplicationContext(ctx)
	if err != nil {
		return nil, err
	}

	subject := ext.prepareSubject(appCtx.Tenant, appCtx.Group, appCtx.Application)

	apiHosts, ok := ctx.Value(APIHostsKey).(APIHosts)
	if !ok {
		return newClientCertificateContext(appCtx, subject), nil
	}

	metadataURL := ""
	eventsURL := ""

	if apiHosts.MetadataHost != "" {
		metadataURL = fmt.Sprintf(MetadataURLFormat, apiHosts.MetadataHost, appCtx.GetApplication())
	}
	if apiHosts.EventsHost != "" {
		eventsURL = fmt.Sprintf(EventsURLFormat, apiHosts.EventsHost, appCtx.GetApplication())
	}

	extendedCtx := &ExtendedApplicationContext{
		ApplicationContext: appCtx,
		RuntimeURLs: RuntimeURLs{
			MetadataURL: metadataURL,
			EventsURL:   eventsURL,
		},
	}

	return newClientCertificateContext(extendedCtx, subject), nil
}

func (ext *ContextExtractor) prepareSubject(tenant, group, commonName string) string {
	organization := tenant
	organizationalUnit := group

	if isEmpty(organization) {
		organization = ext.subjectDefaults.Organization
	}

	if isEmpty(organizationalUnit) {
		organizationalUnit = ext.subjectDefaults.OrganizationalUnit
	}

	return fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", organizationalUnit, organization, ext.subjectDefaults.Locality, ext.subjectDefaults.Province, ext.subjectDefaults.Country, commonName)
}

func (ext *ContextExtractor) CreateClusterClientContextService(ctx context.Context) (ClientContextService, apperrors.AppError) {
	clusterCtx, ok := ctx.Value(ClusterContextKey).(ClusterContext)
	if !ok {
		return nil, apperrors.Internal("Failed to create params when reading ClusterContext")
	}

	subject := ext.prepareSubject(clusterCtx.Tenant, clusterCtx.Group, RuntimeDefaultCommonName)

	return newClientCertificateContext(clusterCtx, subject), nil
}

func ExtractApplicationContext(ctx context.Context) (ApplicationContext, apperrors.AppError) {
	appCtx, ok := ctx.Value(ApplicationContextKey).(ApplicationContext)
	if !ok {
		return ApplicationContext{}, apperrors.Internal("Failed to create params when reading ApplicationContext")
	}
	return appCtx, nil
}

func isEmpty(str string) bool {
	return str == ""
}

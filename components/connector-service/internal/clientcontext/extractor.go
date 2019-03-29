package clientcontext

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

const (
	MetadataURLFormat = "https://%s/%s/v1/metadata/services"
	EventsURLFormat   = "https://%s/%s/v1/events"
)

type ConnectorClientExtractor func(ctx context.Context) (ClientContextService, apperrors.AppError)

type ApplicationContextExtractor func(ctx context.Context) (ApplicationContext, apperrors.AppError)

func CreateApplicationClientContextService(ctx context.Context) (ClientContextService, apperrors.AppError) {
	appCtx, err := ExtractApplicationContext(ctx)

	if err != nil {
		return nil, err
	}

	apiHosts, ok := ctx.Value(APIHostsKey).(APIHosts)
	if !ok {
		return appCtx, nil
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

	return extendedCtx, nil
}

func CreateClusterClientContextService(ctx context.Context) (ClientContextService, apperrors.AppError) {
	clusterCtx, ok := ctx.Value(ClusterContextKey).(ClusterContext)
	if !ok {
		return nil, apperrors.Internal("Failed to create params when reading ClusterContext")
	}

	return clusterCtx, nil
}

func ExtractApplicationContext(ctx context.Context) (ApplicationContext, apperrors.AppError) {
	appCtx, ok := ctx.Value(ApplicationContextKey).(ApplicationContext)
	if !ok {
		return ApplicationContext{}, apperrors.Internal("Failed to create params when reading ApplicationContext")
	}
	return appCtx, nil
}

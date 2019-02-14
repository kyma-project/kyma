package clientcontext

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

const (
	MetadataURLFormat = "https://%s/%s/v1/metadata/services"
	EventsURLFormat   = "https://%s/%s/v1/events"
)

type ConnectorClientExtractor func(ctx context.Context) (ConnectorClientContext, apperrors.AppError)

func ExtractApplicationContext(ctx context.Context) (ConnectorClientContext, apperrors.AppError) {
	appCtx, ok := ctx.Value(ApplicationContextKey).(ApplicationContext)
	if !ok {
		return nil, apperrors.Internal("Failed to create params when reading ApplicationContext")
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

func ExtractClusterContext(ctx context.Context) (ConnectorClientContext, apperrors.AppError) {
	clusterCtx, ok := ctx.Value(ClusterContextKey).(ClusterContext)
	if !ok {
		return nil, apperrors.Internal("Failed to create params when reading ClusterContext")
	}

	return clusterCtx, nil
}

func EmptyClusterContext(_ context.Context) (ConnectorClientContext, apperrors.AppError) {
	return &ClusterContext{}, nil
}

func NewClusterContextExtender(token string, tokenResolver tokens.Resolver) (ContextExtender, apperrors.AppError) {
	var clusterContext ClusterContext
	err := tokenResolver.Resolve(token, &clusterContext)

	return clusterContext, err
}

func NewApplicationContextExtender(token string, tokenResolver tokens.Resolver) (ContextExtender, apperrors.AppError) {
	var appContext ApplicationContext
	err := tokenResolver.Resolve(token, &appContext)

	return appContext, err
}

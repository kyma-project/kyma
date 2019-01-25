package clientcontext

import (
	"context"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

type ConnectorClientExtractor func(ctx context.Context) (ConnectorClientContext, apperrors.AppError)

func ExtractApplicationContext(ctx context.Context) (ConnectorClientContext, apperrors.AppError) {
	appCtx, ok := ctx.Value(ApplicationContextKey).(ApplicationContext)
	if !ok {
		return nil, apperrors.Internal("Failed to create params when reading ApplicationContext")
	}

	return appCtx, nil
}

func ExtractClusterContext(ctx context.Context) (ConnectorClientContext, apperrors.AppError) {
	clusterCtx, ok := ctx.Value(ClusterContextKey).(ClusterContext)
	if !ok {
		return nil, apperrors.Internal("Failed to create params when reading ClusterContext")
	}

	return clusterCtx, nil
}

func ResolveClusterContextExtender(token string, tokenResolver tokens.Resolver) (ContextExtender, apperrors.AppError) {
	var clusterContext ClusterContext
	err := tokenResolver.Resolve(token, &clusterContext)

	return clusterContext, err
}

func ResolveApplicationContextExtender(token string, tokenResolver tokens.Resolver) (ContextExtender, apperrors.AppError) {
	var appContext ApplicationContext
	err := tokenResolver.Resolve(token, &appContext)

	return appContext, err
}

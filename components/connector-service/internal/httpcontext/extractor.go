package httpcontext

import (
	"context"

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

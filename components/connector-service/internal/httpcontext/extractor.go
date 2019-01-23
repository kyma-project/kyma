package httpcontext

import (
	"context"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

type SerializerExtractor func(ctx context.Context) (Serializer, apperrors.AppError)

func ExtractSerializableApplicationContext(ctx context.Context) (Serializer, apperrors.AppError) {
	appCtx, ok := ctx.Value(ApplicationContextKey).(ApplicationContext)
	if !ok {
		return nil, apperrors.Internal("Failed to create params when reading ApplicationContext")
	}

	return appCtx, nil
}

func ExtractSerializableClusterContext(ctx context.Context) (Serializer, apperrors.AppError) {
	clusterCtx, ok := ctx.Value(ClusterContextKey).(ClusterContext)
	if !ok {
		return nil, apperrors.Internal("Failed to create params when reading ClusterContext")
	}

	return clusterCtx, nil
}

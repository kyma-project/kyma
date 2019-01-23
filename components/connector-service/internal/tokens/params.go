package tokens

import (
	"context"
	"encoding/json"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/middlewares"
)

type TokenParamsParser func(ctx context.Context) (TokenParams, apperrors.AppError)

type ApplicationTokenParams struct {
	ClusterTokenParams
	Application string
}

func (params ApplicationTokenParams) ToJSON() ([]byte, error) {
	return json.Marshal(params)
}

type ClusterTokenParams struct {
	Tenant string
	Group  string
}

func (params ClusterTokenParams) ToJSON() ([]byte, error) {
	return json.Marshal(params)
}

func NewApplicationTokenParams(ctx context.Context) (TokenParams, apperrors.AppError) {
	appCtx, ok := ctx.Value(middlewares.ApplicationContextKey).(middlewares.ApplicationContext)
	if !ok {
		return nil, apperrors.Internal("Failed to create params when reading ApplicationContext")
	}

	clusterCtx, ok := ctx.Value(middlewares.ClusterContextKey).(middlewares.ClusterContext)
	if !ok {
		return nil, apperrors.Internal("Failed to create params when reading ClusterContext")
	}

	return ApplicationTokenParams{
		Application: appCtx.Application,
		ClusterTokenParams: ClusterTokenParams{
			Tenant: clusterCtx.Tenant,
			Group:  clusterCtx.Group,
		},
	}, nil
}

func NewClusterTokenParams(ctx context.Context) (TokenParams, apperrors.AppError) {
	clusterCtx, ok := ctx.Value(middlewares.ClusterContextKey).(middlewares.ClusterContext)
	if !ok {
		return nil, apperrors.Internal("Failed to create params when reading ClusterContext")
	}

	return ClusterTokenParams{
		Tenant: clusterCtx.Tenant,
		Group:  clusterCtx.Group,
	}, nil
}

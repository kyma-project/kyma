package tokens

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
)

type Resolver interface {
	Resolve(token string, destination interface{}) apperrors.AppError
}

type tokenResolver struct {
	store tokencache.TokenCache
}

func NewTokenResolver(store tokencache.TokenCache) *tokenResolver {
	return &tokenResolver{
		store: store,
	}
}

func (svc *tokenResolver) Resolve(token string, destination interface{}) apperrors.AppError {
	encodedParams, found := svc.store.Get(token)
	if !found {
		return apperrors.NotFound("Token not found")
	}

	err := json.Unmarshal([]byte(encodedParams), destination)
	if err != nil {
		return apperrors.Internal("Failed to unmarshal token params, %s", err.Error())
	}

	return nil
}

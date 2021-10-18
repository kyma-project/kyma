package tokens

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/tokens/tokencache"
)

type Manager interface {
	Resolve(token string, destination interface{}) apperrors.AppError
	Delete(token string)
}

type tokenManager struct {
	store tokencache.TokenCache
}

func NewTokenManager(store tokencache.TokenCache) *tokenManager {
	return &tokenManager{
		store: store,
	}
}

func (svc *tokenManager) Resolve(token string, destination interface{}) apperrors.AppError {
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

func (svc *tokenManager) Delete(token string) {
	svc.store.Delete(token)
}

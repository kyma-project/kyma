package tokens

import (
	"time"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
)

type Generator func() (string, apperrors.AppError)

type Serializer interface {
	ToJSON() ([]byte, error)
}

type Manager interface {
	Save(serializableContext Serializer) (string, apperrors.AppError)
	Replace(token string, serializableContext Serializer) (string, apperrors.AppError)
	Delete(token string)
}

type tokenManager struct {
	tokenTTL  time.Duration
	store     tokencache.TokenCache
	generator Generator
}

func NewTokenManager(tokenTTL time.Duration, store tokencache.TokenCache, generator Generator) *tokenManager {
	return &tokenManager{
		tokenTTL:  tokenTTL,
		store:     store,
		generator: generator,
	}
}

func (svc *tokenManager) Save(serializableContext Serializer) (string, apperrors.AppError) {
	jsonData, err := serializableContext.ToJSON()
	if err != nil {
		return "", apperrors.Internal("Failed to serialize token params to JSON, %s", err.Error())
	}

	token, err := svc.generator()
	if err != nil {
		return "", apperrors.Internal("Failed to generate token, %s", err.Error())
	}

	svc.store.Put(token, string(jsonData), svc.tokenTTL)

	return token, nil
}

func (svc *tokenManager) Replace(token string, serializableContext Serializer) (string, apperrors.AppError) {
	svc.store.Delete(token)

	return svc.Save(serializableContext)
}

func (svc *tokenManager) Delete(token string) {
	svc.store.Delete(token)
}

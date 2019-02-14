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

type Creator interface {
	Save(serializableContext Serializer) (string, apperrors.AppError)
}

type tokenCreator struct {
	tokenTTL  time.Duration
	store     tokencache.TokenCache
	generator Generator
}

func NewTokenCreator(tokenTTL time.Duration, store tokencache.TokenCache, generator Generator) *tokenCreator {
	return &tokenCreator{
		tokenTTL:  tokenTTL,
		store:     store,
		generator: generator,
	}
}

func (svc *tokenCreator) Save(serializableContext Serializer) (string, apperrors.AppError) {
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

package tokens

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
)

type Generator func() (string, apperrors.AppError)

type Serializer interface {
	ToJSON() ([]byte, error)
}

type Service interface {
	Save(serializableContext Serializer) (string, apperrors.AppError)
	Replace(token string, serializableContext Serializer) (string, apperrors.AppError)
	Resolve(token string, destination interface{}) apperrors.AppError
	Delete(token string)
}

type Creator interface {
	Save(serializableContext Serializer) (string, apperrors.AppError)
	Replace(token string, serializableContext Serializer) (string, apperrors.AppError)
}

type Remover interface {
	Delete(token string)
}

type Resolver interface {
	Resolve(token string, destination interface{}) apperrors.AppError
}

type tokenService struct {
	store     tokencache.TokenCache
	generator Generator
}

func NewTokenService(store tokencache.TokenCache, generator Generator) *tokenService {
	return &tokenService{
		store:     store,
		generator: generator,
	}
}

func (svc *tokenService) Save(serializableContext Serializer) (string, apperrors.AppError) {
	jsonData, err := serializableContext.ToJSON()
	if err != nil {
		return "", apperrors.Internal("Faild to serilize token params to JSON, %s", err.Error())
	}

	token, err := svc.generator()
	if err != nil {
		return "", apperrors.Internal("Failed to generate token, %s", err.Error())
	}

	svc.store.Put(token, string(jsonData))

	return token, nil
}

func (svc *tokenService) Replace(token string, serializableContext Serializer) (string, apperrors.AppError) {
	svc.store.Delete(token)

	return svc.Save(serializableContext)
}

func (svc *tokenService) Delete(token string) {
	svc.store.Delete(token)
}

func (svc *tokenService) Resolve(token string, destination interface{}) apperrors.AppError {
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

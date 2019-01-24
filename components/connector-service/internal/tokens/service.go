package tokens

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/connector-service/internal/httpcontext"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
)

type Generator func() (string, apperrors.AppError)

type Service interface {
	Save(serializableContext httpcontext.Serializer) (string, apperrors.AppError)
	Replace(token string, serializableContext httpcontext.Serializer) (string, apperrors.AppError)
	Resolve(token string, destination interface{}) apperrors.AppError
	Delete(token string)
}

type Creator interface {
	Save(serializableContext httpcontext.Serializer) (string, apperrors.AppError)
	Replace(token string, serializableContext httpcontext.Serializer) (string, apperrors.AppError)
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

func (svc *tokenService) Save(serializableContext httpcontext.Serializer) (string, apperrors.AppError) {
	jsonData, err := serializableContext.ToJSON()
	if err != nil {
		return "", apperrors.Internal("Faild to serilize token params to JSON, %s", err.Error())
	}

	token, err := svc.generator()
	if err != nil {
		return "", apperrors.Internal("Failed to generate token, %s", err.Error())
	}

	svc.store.Put(string(jsonData), token)

	return token, nil
}

func (svc *tokenService) Replace(token string, serializableContext httpcontext.Serializer) (string, apperrors.AppError) {
	// TODO
	return "", nil
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

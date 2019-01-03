package tokens

import (
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

type Cache interface {
	Set(string, interface{}, time.Duration)
	Get(string) (interface{}, bool)
	Delete(string)
}

type Service interface {
	CreateAppToken(identifier string, data *TokenData) (string, apperrors.AppError)
	GetAppToken(identifier string) (*TokenData, bool)
	DeleteAppToken(identifier string)
	CreateClusterToken(identifier string) (string, apperrors.AppError)
	GetClusterToken(identifier string) (string, bool)
	DeleteClusterToken(identifier string)
}

type ApplicationService interface {
	CreateAppToken(identifier string, data *TokenData) (string, apperrors.AppError)
	GetAppToken(identifier string) (*TokenData, bool)
	DeleteAppToken(identifier string)
}

type ClusterService interface {
	CreateClusterToken(identifier string) (string, apperrors.AppError)
	GetClusterToken(identifier string) (string, bool)
	DeleteClusterToken(identifier string)
}

type tokenService struct {
	tokenLength       int
	generatorFunc     func(length int) (string, apperrors.AppError)
	appTokenCache     Cache
	clusterTokenCache Cache
}

func NewTokenService(tokenLength int, generatorFunc func(length int) (string, apperrors.AppError), appTokenCache Cache, clusterTokenCache Cache) Service {
	return &tokenService{
		tokenLength:       tokenLength,
		generatorFunc:     generatorFunc,
		appTokenCache:     appTokenCache,
		clusterTokenCache: clusterTokenCache,
	}
}

func (ts *tokenService) CreateAppToken(identifier string, tokenData *TokenData) (string, apperrors.AppError) {
	token, err := ts.generatorFunc(ts.tokenLength)
	if err != nil {
		return "", err
	}
	tokenData.Token = token

	ts.appTokenCache.Set(identifier, tokenData, cache.DefaultExpiration)
	return token, nil
}

func (ts *tokenService) GetAppToken(identifier string) (*TokenData, bool) {
	token, found := ts.appTokenCache.Get(identifier)
	if !found {
		return nil, found
	}

	return token.(*TokenData), found
}

func (ts *tokenService) DeleteAppToken(identifier string) {
	ts.appTokenCache.Delete(identifier)
}

func (ts *tokenService) CreateClusterToken(identifier string) (string, apperrors.AppError) {
	token, err := ts.generatorFunc(ts.tokenLength)
	if err != nil {
		return "", err
	}

	ts.clusterTokenCache.Set(identifier, token, cache.DefaultExpiration)
	return token, nil
}

func (ts *tokenService) GetClusterToken(identifier string) (string, bool) {
	token, found := ts.clusterTokenCache.Get(identifier)
	if !found {
		return "", found
	}

	return token.(string), found
}

func (ts *tokenService) DeleteClusterToken(identifier string) {
	ts.clusterTokenCache.Delete(identifier)
}

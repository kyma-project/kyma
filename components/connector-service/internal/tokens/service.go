package tokens

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

type Service interface {
	CreateToken(app string, data *TokenData) (string, apperrors.AppError)
	GetToken(identifier string) (*TokenData, bool)
	DeleteToken(identifier string)
}

type tokenService struct {
	tokenLength int
	tokenCache  Cache
}

func NewTokenService(tokenLength int, tokenCache Cache) Service {
	return &tokenService{tokenLength: tokenLength, tokenCache: tokenCache}
}

func (ts *tokenService) CreateToken(app string, tokenData *TokenData) (string, apperrors.AppError) {
	token, err := generateRandomString(ts.tokenLength)
	if err != nil {
		return "", err
	}
	tokenData.Token = token

	ts.tokenCache.Put(app, tokenData)
	return token, nil
}

func generateRandomBytes(number int) ([]byte, apperrors.AppError) {
	bytes := make([]byte, number)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, apperrors.Internal("Failed to generate random bytes: %s", err)
	}

	return bytes, nil
}

func generateRandomString(length int) (string, apperrors.AppError) {
	bytes, err := generateRandomBytes(length)
	return base64.URLEncoding.EncodeToString(bytes), err
}

func (ts *tokenService) GetToken(identifier string) (*TokenData, bool) {
	return ts.tokenCache.Get(identifier)
}

func (ts *tokenService) DeleteToken(identifier string) {
	ts.tokenCache.Delete(identifier)
}

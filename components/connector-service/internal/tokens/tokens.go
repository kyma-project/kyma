package tokens

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
)

type TokenGenerator interface {
	NewToken(app string) (string, apperrors.AppError)
}

type tokenGenerator struct {
	tokenLength int
	tokenCache  tokencache.TokenCache
}

func NewTokenGenerator(tokenLength int, tokenCache tokencache.TokenCache) TokenGenerator {
	return &tokenGenerator{tokenLength: tokenLength, tokenCache: tokenCache}
}

func (tg *tokenGenerator) NewToken(app string) (string, apperrors.AppError) {
	token, err := generateRandomString(tg.tokenLength)
	if err != nil {
		return "", err
	}

	tg.tokenCache.Put(app, token)
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

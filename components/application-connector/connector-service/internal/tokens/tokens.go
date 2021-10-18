package tokens

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/apperrors"
)

type TokenGenerator interface {
	NewToken() (string, apperrors.AppError)
}

type tokenGenerator struct {
	tokenLength int
}

func NewTokenGenerator(tokenLength int) TokenGenerator {
	return &tokenGenerator{tokenLength: tokenLength}
}

func (tg *tokenGenerator) NewToken() (string, apperrors.AppError) {
	return generateRandomString(tg.tokenLength)
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

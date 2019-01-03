package tokens

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

func GenerateRandomString(length int) (string, apperrors.AppError) {
	bytes, err := generateRandomBytes(length)
	return base64.URLEncoding.EncodeToString(bytes), err
}

func generateRandomBytes(number int) ([]byte, apperrors.AppError) {
	bytes := make([]byte, number)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, apperrors.Internal("Failed to generate random bytes: %s", err)
	}

	return bytes, nil
}

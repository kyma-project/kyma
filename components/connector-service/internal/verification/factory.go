package verification

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

// TODO - change the service name to something more suitable

const (
	defaultGroup  = "default"
	defaultTenant = "default"
)

type Service interface {
	// Verify the entity for with the certificate should be created
	Verify(request *http.Request, identifier string) (*tokens.TokenData, apperrors.AppError)
}

func NewVerificationService(globalMode bool) Service {
	if globalMode {
		return newMultiTenantIdentificationService()
	}

	return newBasicIdentificationService()
}

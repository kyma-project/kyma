package verification

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

type basic struct {
}

func newBasicIdentificationService() Service {
	return &basic{}
}

func (svc *basic) Verify(request *http.Request, identifier string) (*tokens.TokenData, apperrors.AppError) {

	return &tokens.TokenData{
		Group:  defaultGroup,
		Tenant: defaultTenant,
	}, nil
}

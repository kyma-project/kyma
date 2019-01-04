package verification

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

type multiTenant struct {
}

func newMultiTenantIdentificationService() Service {
	return &multiTenant{}
}

func (svc *multiTenant) Verify(request *http.Request, identifier string) (*tokens.TokenData, apperrors.AppError) {
	group := request.Header.Get("Group")
	if group == "" {
		group = defaultGroup
	}

	tenant := request.Header.Get("Tenant")
	if tenant == "" {
		return nil, apperrors.BadRequest("Error - tenant not specified")
	}

	return &tokens.TokenData{
		Group:  group,
		Tenant: tenant,
	}, nil
}

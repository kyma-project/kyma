package middlewares

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates/revocationlist"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"net/http"
)

type revocationCheckMiddleware struct {
	revocationList revocationlist.RevocationListRepository
}

func NewRevocationCheckMiddleware(revocationList revocationlist.RevocationListRepository) *revocationCheckMiddleware {
	return &revocationCheckMiddleware{
		revocationList: revocationList,
	}
}

func (rcm revocationCheckMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		certificate := r.Header.Get(externalapi.CertificateHeader)

		hash, e := certificates.CalculateHash(certificate)

		if e != nil {
			httphelpers.RespondWithErrorAndLog(w, apperrors.Internal("Failed to calculate hash. Certificate could not be unescaped"))
			return
		}

		contains, err := rcm.revocationList.Contains(hash)

		if err != nil {
			httphelpers.RespondWithErrorAndLog(w, apperrors.Internal("Failed to read revocation list"))
			return
		}

		if contains {
			httphelpers.RespondWithErrorAndLog(w, apperrors.Forbidden("Certificate has been revoked"))
			return
		}

		handler.ServeHTTP(w, r)
	})
}

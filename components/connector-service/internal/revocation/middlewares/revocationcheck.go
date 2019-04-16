package middlewares

import (
	"net/http"
	"net/url"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/kyma-project/kyma/components/connector-service/internal/revocation"
)

type revocationCheckMiddleware struct {
	revocationList revocation.RevocationListRepository
}

func NewRevocationCheckMiddleware(revocationList revocation.RevocationListRepository) *revocationCheckMiddleware {
	return &revocationCheckMiddleware{
		revocationList: revocationList,
	}
}

func (rcm revocationCheckMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		certificate := r.Header.Get(externalapi.CertificateHeader)

		pemCert, err := url.PathUnescape(certificate)
		if err != nil {
			httphelpers.RespondWithErrorAndLog(w, apperrors.Internal("Failed to unescape characters from certificate."))
			return
		}

		hash, err := certificates.FingerprintSHA256([]byte(pemCert))
		if err != nil {
			httphelpers.RespondWithErrorAndLog(w, apperrors.Internal("Failed to calculate certificate hash."))
			return
		}

		contains, err := rcm.revocationList.Contains(hash)
		if err != nil {
			httphelpers.RespondWithErrorAndLog(w, apperrors.Internal("Failed to read revocation list."))
			return
		}

		if contains {
			httphelpers.RespondWithErrorAndLog(w, apperrors.Forbidden("Certificate has been revoked."))
			return
		}

		handler.ServeHTTP(w, r)
	})
}

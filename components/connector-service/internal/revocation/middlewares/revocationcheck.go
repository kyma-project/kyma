package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/kyma-project/kyma/components/connector-service/internal/revocation"
)

type revocationCheckMiddleware struct {
	revocationList revocation.RevocationListRepository
	headerParser   certificates.HeaderParser
}

func NewRevocationCheckMiddleware(revocationList revocation.RevocationListRepository, headerParser certificates.HeaderParser) *revocationCheckMiddleware {
	return &revocationCheckMiddleware{
		revocationList: revocationList,
		headerParser:   headerParser,
	}
}

func (rcm revocationCheckMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		certInfo, appError := rcm.headerParser.ParseCertificateHeader(*r)

		if appError != nil {
			httphelpers.RespondWithErrorAndLog(w, appError)
			return
		}

		contains, err := rcm.revocationList.Contains(certInfo.Hash)
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

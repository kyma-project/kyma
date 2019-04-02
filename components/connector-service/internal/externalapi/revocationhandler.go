package externalapi

import (
	"net/http"
	"net/url"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/kyma-project/kyma/components/connector-service/internal/revocation"
)

const (
	CertificateHeader = "Client-Certificate"
)

type revocationHandler struct {
	revocationList revocation.RevocationListRepository
}

func NewRevocationHandler(revocationList revocation.RevocationListRepository) *revocationHandler {
	return &revocationHandler{
		revocationList: revocationList,
	}
}

func (handler revocationHandler) Revoke(w http.ResponseWriter, r *http.Request) {

	hash, appError := handler.getCertificateHash(r)
	if appError != nil {
		httphelpers.RespondWithErrorAndLog(w, appError)
		return
	}

	appError = handler.addToRevocationList(hash)
	if appError != nil {
		httphelpers.RespondWithErrorAndLog(w, appError)
		return
	}

	httphelpers.Respond(w, http.StatusCreated)
}

func (handler revocationHandler) getCertificateHash(r *http.Request) (string, apperrors.AppError) {
	cert := r.Header.Get(CertificateHeader)

	if cert == "" {
		return "", apperrors.Internal("Cannot calculate certificate hash. Certificate not passed to the service.")
	}

	pemCert, err := url.PathUnescape(cert)
	if err != nil {
		return "", apperrors.Internal("Failed to unescape characters from certificate.")
	}

	hash, err := certificates.FingerprintSHA256([]byte(pemCert))
	if err != nil {
		return "", apperrors.Internal("Failed to calculate certificate hash.")
	}

	return hash, nil
}

func (handler revocationHandler) addToRevocationList(hash string) apperrors.AppError {
	err := handler.revocationList.Insert(hash)
	if err != nil {
		return apperrors.Internal("Unable to mark certificate as revoked: %s.", err)
	}

	return nil
}

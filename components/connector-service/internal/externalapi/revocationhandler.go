package externalapi

import (
	"context"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/kyma-project/kyma/components/connector-service/internal/revocation"
)

type revocationHandler struct {
	ctx            context.Context
	revocationList revocation.RevocationListRepository
	headerParser   certificates.HeaderParser
}

func NewRevocationHandler(revocationList revocation.RevocationListRepository, headerParser certificates.HeaderParser) *revocationHandler {
	return &revocationHandler{
		ctx:            context.Background(),
		revocationList: revocationList,
		headerParser:   headerParser,
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
	certInfo, appError := handler.headerParser.ParseCertificateHeader(*r)
	if appError != nil {
		return "", appError
	}

	return certInfo.Hash, nil
}

func (handler revocationHandler) addToRevocationList(hash string) apperrors.AppError {
	err := handler.revocationList.Insert(handler.ctx, hash)
	if err != nil {
		return apperrors.Internal("Unable to mark certificate as revoked: %s.", err)
	}

	return nil
}

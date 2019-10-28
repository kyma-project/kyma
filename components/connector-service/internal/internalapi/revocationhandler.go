package internalapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/kyma-project/kyma/components/connector-service/internal/revocation"
	"github.com/sirupsen/logrus"
)

type revocationBody struct {
	Hash string
}

type revocationHandler struct {
	revocationList revocation.RevocationListRepository
}

func NewRevocationHandler(revocationListRepository revocation.RevocationListRepository) *revocationHandler {
	return &revocationHandler{
		revocationList: revocationListRepository,
	}
}

func (handler revocationHandler) Revoke(w http.ResponseWriter, request *http.Request) {
	rb, appError := handler.readBody(request)
	if appError != nil {
		httphelpers.RespondWithErrorAndLog(w, appError)
		return
	}

	appError = handler.addToRevocationList(rb.Hash)
	if appError != nil {
		httphelpers.RespondWithErrorAndLog(w, appError)
		return
	}

	httphelpers.Respond(w, http.StatusCreated)
}

func (handler revocationHandler) readBody(request *http.Request) (*revocationBody, apperrors.AppError) {
	var rb revocationBody

	b, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, apperrors.BadRequest("Error while reading request body: %s.", err)
	}
	defer request.Body.Close()

	err = json.Unmarshal(b, &rb)
	if err != nil {
		return nil, apperrors.BadRequest("Error while unmarshalling request body: %s.", err)
	}

	if rb.Hash == "" {
		return nil, apperrors.BadRequest("Error while unmarshalling request body: certificate hash value not provided.")
	}

	return &rb, nil
}

func (handler revocationHandler) addToRevocationList(hash string) apperrors.AppError {
	err := handler.revocationList.Insert(hash)

	logrus.Warningf("Adding certificate with hash: %s to revocation list.", hash)
	if err != nil {
		return apperrors.Internal("Unable to mark certificate as revoked: %s.", err)
	}

	return nil
}

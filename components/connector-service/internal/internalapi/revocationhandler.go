package internalapi

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates/revocationlist"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"io/ioutil"
	"net/http"
)

type revocationBody struct {
	Hash string
}

type revocationHandler struct {
	revocationList revocationlist.RevocationListRepository
}

func NewRevocationHandler(revocationListRepository revocationlist.RevocationListRepository) *revocationHandler{
	return &revocationHandler{
		revocationList: revocationListRepository,
	}
}

func (handler revocationHandler) Revoke(w http.ResponseWriter, request *http.Request) {
	rb, appError := handler.readRevocationBody(request)

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

func (handler revocationHandler) readRevocationBody(request *http.Request) (*revocationBody, apperrors.AppError){
	var rb revocationBody

	b, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, apperrors.BadRequest("Error while reading request body: %s", err)
	}
	defer request.Body.Close()

	err = json.Unmarshal(b, &rb)
	if err != nil {
		return nil, apperrors.BadRequest("Error while unmarshalling request body: %s", err)
	}

	if rb.Hash == ""{
		return nil, apperrors.BadRequest("Error while unmarshalling request body: hash value not provided")
	}

	return &rb, nil
}

func (handler revocationHandler) addToRevocationList(hash string) apperrors.AppError {
	err := handler.revocationList.Insert(hash)

	if err != nil {
		return apperrors.Internal("Unable to mark certificate as revoked: %s.", err)
	}

	return nil
}


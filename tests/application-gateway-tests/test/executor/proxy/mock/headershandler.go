package mock

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

type queryParamsHandler struct {
	logger *log.Entry
}

func NewQueryParamsHandler() *queryParamsHandler {
	return &queryParamsHandler{
		logger: log.WithField("Handler", "Headers"),
	}
}

func (qph *queryParamsHandler) RequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	expectedParam := vars["param"]
	expectedParamValue := vars["value"]
	qph.logger.Infof("Handling request. Expected: param: %s, with value: %s", expectedParam, expectedParamValue)

	paramValue := r.URL.Query().Get(expectedParam)

	if expectedParamValue != paramValue {
		qph.logger.Errorf("Invalid query parameter value provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	successResponse(w)
}

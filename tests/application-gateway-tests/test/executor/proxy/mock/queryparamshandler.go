package mock

import (
	"net/http"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

type queryParamsHandler struct {
	logger *log.Entry
}

func NewQueryParamsHandler() *queryParamsHandler {
	return &queryParamsHandler{
		logger: log.WithField("Handler", "QueryParams"),
	}
}

func (qph *queryParamsHandler) QueryParamsHandler(w http.ResponseWriter, r *http.Request) {
	err := qph.checkQueryParams(r)
	if err != nil {
		qph.logger.Errorf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (qph *queryParamsHandler) QueryParamsHandlerSpec(w http.ResponseWriter, r *http.Request) {
	err := qph.checkQueryParams(r)
	if err != nil {
		qph.logger.Errorf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, "spec.json")
}

func (qph *queryParamsHandler) checkQueryParams(r *http.Request) error {
	vars := mux.Vars(r)
	expectedParam := vars["param"]
	expectedParamValue := vars["value"]
	qph.logger.Infof("Handling request. Expected: param: %s, with value: %s", expectedParam, expectedParamValue)

	paramValue := r.URL.Query().Get(expectedParam)

	if expectedParamValue != paramValue {
		return errors.New("Invalid query parameter value provided")
	}

	return nil
}

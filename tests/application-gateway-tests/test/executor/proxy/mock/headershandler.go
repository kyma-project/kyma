package mock

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

type headersHandler struct {
	logger *log.Entry
}

func NewHeadersHandler() *headersHandler {
	return &headersHandler{
		logger: log.WithField("Handler", "Headers"),
	}
}

func (h *headersHandler) RequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	expectedHeader := vars["header"]
	expectedHeaderValue := vars["value"]
	h.logger.Infof("Handling request. Expected: param: %s, with value: %s", expectedHeader, expectedHeaderValue)

	headerValue := r.Header.Get(expectedHeader)

	if expectedHeaderValue != headerValue {
		h.logger.Errorf("Invalid header value provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	successResponse(w)
}

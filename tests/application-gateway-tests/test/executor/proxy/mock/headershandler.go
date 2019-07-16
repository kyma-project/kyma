package mock

import (
	"net/http"

	"github.com/pkg/errors"

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

func (h *headersHandler) HeadersHandler(w http.ResponseWriter, r *http.Request) {
	err := h.checkCustomHeaders(r)
	if err != nil {
		h.logger.Errorf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *headersHandler) HeadersHandlerSpec(w http.ResponseWriter, r *http.Request) {
	err := h.checkCustomHeaders(r)
	if err != nil {
		h.logger.Errorf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, "spec.json")
}

func (h *headersHandler) checkCustomHeaders(r *http.Request) error {
	vars := mux.Vars(r)
	expectedHeader := vars["header"]
	expectedHeaderValue := vars["value"]
	h.logger.Infof("Handling request. Expected: param: %s, with value: %s", expectedHeader, expectedHeaderValue)

	headerValue := r.Header.Get(expectedHeader)

	if expectedHeaderValue != headerValue {
		return errors.New("Invalid header value provided")
	}

	return nil
}

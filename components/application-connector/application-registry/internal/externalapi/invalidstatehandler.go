package externalapi

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/httperrors"
	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/httptools"
)

type invalidStateHandler struct {
	Message string
}

func NewInvalidStateMetadataHandler(message string) MetadataHandler {
	return &invalidStateHandler{Message: message}
}

func (ish *invalidStateHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLogger(r)
	httptools.DumpRequestToLog(r, contextLogger)

	ish.HandleRequest(w, r)
}

func (ish *invalidStateHandler) GetService(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLoggerWithId(r)
	httptools.DumpRequestToLog(r, contextLogger)

	ish.HandleRequest(w, r)
}

func (ish *invalidStateHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLogger(r)
	httptools.DumpRequestToLog(r, contextLogger)

	ish.HandleRequest(w, r)
}

func (ish *invalidStateHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLoggerWithId(r)
	httptools.DumpRequestToLog(r, contextLogger)

	ish.HandleRequest(w, r)
}

func (ish *invalidStateHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLoggerWithId(r)
	httptools.DumpRequestToLog(r, contextLogger)

	ish.HandleRequest(w, r)
}

func (ish *invalidStateHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLogger(r)
	contextLogger.Errorf("Error handling request: %s.", ish.Message)

	statusCode := http.StatusInternalServerError
	responseBody := httperrors.ErrorResponse{Code: statusCode, Error: ish.Message}

	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(responseBody)
}

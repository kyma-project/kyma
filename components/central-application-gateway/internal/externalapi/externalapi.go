package externalapi

import (
	"go.uber.org/zap"
	"net/http"

	"github.com/gorilla/mux"
)

func NewHandler(lvl zap.AtomicLevel) http.Handler {
	router := mux.NewRouter()

	router.Path("/v1/health").Handler(NewHealthCheckHandler()).Methods(http.MethodGet)
	router.Path("/v1/loglevel").Handler(lvl).Methods(http.MethodGet, http.MethodPut)

	router.NotFoundHandler = NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(405, "Method not allowed.")

	return router
}

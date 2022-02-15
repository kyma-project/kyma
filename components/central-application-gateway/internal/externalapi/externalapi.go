package externalapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewHandler() http.Handler {
	router := mux.NewRouter()

	router.Path("/v1/health").Handler(NewHealthCheckHandler()).Methods(http.MethodGet)

	router.NotFoundHandler = NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(405, "Method not allowed.")

	return router
}

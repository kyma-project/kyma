package externalapi

import (
	"github.com/gorilla/mux"
	"net/http"
)

func NewHandler() http.Handler {

	router := mux.NewRouter()

	router.Path("/v1/health").Handler(NewHealthCheckHandler())

	return router
}

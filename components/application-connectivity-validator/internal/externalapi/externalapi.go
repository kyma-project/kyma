package externalapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewHandler() http.Handler {

	router := mux.NewRouter()

	router.Path("/v1/health").Handler(NewHealthCheckHandler())

	return router
}

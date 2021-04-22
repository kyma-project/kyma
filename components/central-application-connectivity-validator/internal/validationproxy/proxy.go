package validationproxy

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewHandler(proxyHandler http.Handler) http.Handler {

	router := mux.NewRouter()

	router.PathPrefix("/{application}/").HandlerFunc(proxyHandler.ServeHTTP)

	return router
}

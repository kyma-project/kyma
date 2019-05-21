package validationproxy

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewHandler(proxyHandler ProxyHandler) http.Handler {

	router := mux.NewRouter()

	router.PathPrefix("/{application}/").HandlerFunc(proxyHandler.ProxyAppConnectorRequests)

	return router
}

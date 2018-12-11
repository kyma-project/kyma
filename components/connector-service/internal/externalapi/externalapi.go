package externalapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
)

type SignatureHandler interface {
	SignCSR(w http.ResponseWriter, r *http.Request)
}

type InfoHandler interface {
	GetInfo(w http.ResponseWriter, r *http.Request)
}

const apiSpecPath = "connectorapi.yaml"

func NewHandler(sHandler SignatureHandler, iHandler InfoHandler, middlewares []mux.MiddlewareFunc) http.Handler {
	router := mux.NewRouter()

	for _, middleware := range middlewares {
		router.Use(middleware)
	}

	router.Path("/v1").Handler(http.RedirectHandler("/v1/api.yaml", http.StatusMovedPermanently)).Methods(http.MethodGet)
	router.Path("/v1/api.yaml").Handler(NewStaticFileHandler(apiSpecPath)).Methods(http.MethodGet)

	registrationRouterRE := router.PathPrefix("/v1/remoteenvironments").Subrouter()
	registrationRouterRE.HandleFunc("/{appName}/client-certs", sHandler.SignCSR).Methods(http.MethodPost)
	registrationRouterRE.HandleFunc("/{appName}/info", iHandler.GetInfo).Methods(http.MethodGet)

	registrationRouterAPP := router.PathPrefix("/v1/applications").Subrouter()
	registrationRouterAPP.HandleFunc("/{appName}/client-certs", sHandler.SignCSR).Methods(http.MethodPost)
	registrationRouterAPP.HandleFunc("/{appName}/info", iHandler.GetInfo).Methods(http.MethodGet)

	router.NotFoundHandler = errorhandler.NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = errorhandler.NewErrorHandler(405, "Method not allowed.")

	return router
}

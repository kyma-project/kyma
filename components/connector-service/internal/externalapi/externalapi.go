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

func NewHandler(sHandler SignatureHandler, iHandler InfoHandler) http.Handler {
	router := mux.NewRouter()

	registrationRouter := router.PathPrefix("/v1/remoteenvironments").Subrouter()

	registrationRouter.HandleFunc("/{reName}/client-certs", sHandler.SignCSR).Methods(http.MethodPost)
	registrationRouter.HandleFunc("/{reName}/info", iHandler.GetInfo).Methods(http.MethodGet)

	router.NotFoundHandler = errorhandler.NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = errorhandler.NewErrorHandler(405, "Method not allowed.")

	return router
}

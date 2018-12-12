package internalapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
)

type TokenHandler interface {
	CreateToken(w http.ResponseWriter, r *http.Request)
}

func NewHandler(handler TokenHandler, middlewares []mux.MiddlewareFunc) http.Handler {
	router := mux.NewRouter()

	for _, middleware := range middlewares {
		router.Use(middleware)
	}

	tokenRouter := router.PathPrefix("/v1/applications").Subrouter()
	tokenRouter.HandleFunc("/{reName}/tokens", handler.CreateToken).Methods(http.MethodPost)

	router.NotFoundHandler = errorhandler.NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = errorhandler.NewErrorHandler(405, "Method not allowed.")

	return router
}

package externalapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

type MetadataHandler interface {
	CreateService(w http.ResponseWriter, r *http.Request)
	GetService(w http.ResponseWriter, r *http.Request)
	GetServices(w http.ResponseWriter, r *http.Request)
	UpdateService(w http.ResponseWriter, r *http.Request)
	DeleteService(w http.ResponseWriter, r *http.Request)
}

type RedirectionHandler interface {
	Redirect(w http.ResponseWriter, r *http.Request)
}

const apiSpecPath = "api.yaml"

func NewHandler(handler MetadataHandler, middlewares []mux.MiddlewareFunc) http.Handler {
	router := mux.NewRouter()

	apiSpecRedirectHandler := NewRedirectionHandler("/{application}/v1/metadata/api.yaml", http.StatusMovedPermanently)

	for _, middleware := range middlewares {
		router.Use(middleware)
	}

	router.Path("/v1/health").Handler(NewHealthCheckHandler()).Methods(http.MethodGet)
	router.Path("/{application}/v1/metadata/api.yaml").Handler(NewStaticFileHandler(apiSpecPath)).Methods(http.MethodGet)

	metadataRouter := router.PathPrefix("/{application}/v1/metadata").Subrouter()
	metadataRouter.HandleFunc("", apiSpecRedirectHandler.Redirect)
	metadataRouter.HandleFunc("/services", handler.CreateService).Methods(http.MethodPost)
	metadataRouter.HandleFunc("/services", handler.GetServices).Methods(http.MethodGet)
	metadataRouter.HandleFunc("/services/{serviceId}", handler.GetService).Methods(http.MethodGet)
	metadataRouter.HandleFunc("/services/{serviceId}", handler.UpdateService).Methods(http.MethodPut)
	metadataRouter.HandleFunc("/services/{serviceId}", handler.DeleteService).Methods(http.MethodDelete)

	router.NotFoundHandler = NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(405, "Method not allowed.")

	return router
}

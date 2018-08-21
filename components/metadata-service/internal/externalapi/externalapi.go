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

const apiSpecPath = "./docs/api/api.yaml"

func NewHandler(handler MetadataHandler, middlewares []mux.MiddlewareFunc) http.Handler {
	router := mux.NewRouter()

	for _, middleware := range middlewares {
		router.Use(middleware)
	}

	router.Path("/v1/health").Handler(NewHealthCheckHandler()).Methods(http.MethodGet)

	router.Path("/{remoteEnvironment}/v1/").Handler(http.RedirectHandler("/{remoteEnvironment}/v1/metadataapi.yaml", http.StatusMovedPermanently)).Methods(http.MethodGet)

	router.Path("/{remoteEnvironment}/v1/metadataapi.yaml").Handler(NewStaticFileHandler(apiSpecPath)).Methods(http.MethodGet)

	metadataRouter := router.PathPrefix("/{remoteEnvironment}/v1/metadata").Subrouter()
	metadataRouter.HandleFunc("/services", handler.CreateService).Methods(http.MethodPost)
	metadataRouter.HandleFunc("/services", handler.GetServices).Methods(http.MethodGet)
	metadataRouter.HandleFunc("/services/{serviceId}", handler.GetService).Methods(http.MethodGet)
	metadataRouter.HandleFunc("/services/{serviceId}", handler.UpdateService).Methods(http.MethodPut)
	metadataRouter.HandleFunc("/services/{serviceId}", handler.DeleteService).Methods(http.MethodDelete)

	router.NotFoundHandler = NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(405, "Method not allowed.")

	return router
}

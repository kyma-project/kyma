package externalapi

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
)

type Config struct {
	Middlewares          []mux.MiddlewareFunc
	TokenService         tokens.Service
	ContextExtractor     clientcontext.ConnectorClientExtractor
	CertificateURL       string
	GetInfoURL           string
	ConnectorServiceHost string
	Subject              certificates.CSRSubject
	CertService          certificates.Service
}

type SignatureHandler interface {
	SignCSR(w http.ResponseWriter, r *http.Request)
}

type InfoHandler interface {
	GetCSRInfo(w http.ResponseWriter, r *http.Request)
}

const apiSpecPath = "connectorapi.yaml"

func NewHandler(appHandlerCfg, runtimeHandlerCfg Config, globalMiddlewares []mux.MiddlewareFunc) http.Handler {
	router := mux.NewRouter()

	httphelpers.WithMiddlewares(router, globalMiddlewares)

	router.Path("/v1").Handler(http.RedirectHandler("/v1/api.yaml", http.StatusMovedPermanently)).Methods(http.MethodGet)
	router.Path("/v1/api.yaml").Handler(NewStaticFileHandler(apiSpecPath)).Methods(http.MethodGet)

	applicationInfoHandler := NewCSRInfoHandler(appHandlerCfg.TokenService, appHandlerCfg.ContextExtractor, appHandlerCfg.CertificateURL, runtimeHandlerCfg.GetInfoURL, runtimeHandlerCfg.ConnectorServiceHost, appHandlerCfg.Subject, AppURLFormat)
	applicationManagementInfoHandler := NewManagementInfoHandler()
	applicationSignatureHandler := NewSignatureHandler(appHandlerCfg.TokenService, appHandlerCfg.CertService, appHandlerCfg.ContextExtractor)

	applicationRouter := router.PathPrefix("/v1/applications").Subrouter()
	httphelpers.WithMiddlewares(applicationRouter, appHandlerCfg.Middlewares)
	applicationRouter.HandleFunc("/signingRequests/info", applicationInfoHandler.GetCSRInfo).Methods(http.MethodGet)
	applicationRouter.HandleFunc("/management/info", applicationManagementInfoHandler.GetManagementInfo).Methods(http.MethodGet)
	applicationRouter.HandleFunc("/certificates", applicationSignatureHandler.SignCSR).Methods(http.MethodPost)

	runtimeInfoHandler := NewCSRInfoHandler(runtimeHandlerCfg.TokenService, runtimeHandlerCfg.ContextExtractor, runtimeHandlerCfg.CertificateURL, runtimeHandlerCfg.GetInfoURL, runtimeHandlerCfg.ConnectorServiceHost, runtimeHandlerCfg.Subject, RuntimeURLFormat)
	runtimeManagementInfoHandler := NewManagementInfoHandler()
	runtimeSignatureHandler := NewSignatureHandler(runtimeHandlerCfg.TokenService, runtimeHandlerCfg.CertService, runtimeHandlerCfg.ContextExtractor)

	runtimesRouter := router.PathPrefix("/v1/runtimes").Subrouter()
	httphelpers.WithMiddlewares(runtimesRouter, runtimeHandlerCfg.Middlewares)
	runtimesRouter.HandleFunc("/signingRequests/info", runtimeInfoHandler.GetCSRInfo).Methods(http.MethodGet)
	runtimesRouter.HandleFunc("/management/info", runtimeManagementInfoHandler.GetManagementInfo).Methods(http.MethodGet)
	runtimesRouter.HandleFunc("/certificates", runtimeSignatureHandler.SignCSR).Methods(http.MethodPost)

	router.NotFoundHandler = errorhandler.NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = errorhandler.NewErrorHandler(405, "Method not allowed.")

	return router
}

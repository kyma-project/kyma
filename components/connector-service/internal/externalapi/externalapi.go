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
	TokenCreator     tokens.Creator
	ContextExtractor  clientcontext.ConnectorClientExtractor
	ManagementInfoURL string
	BaseURL           string
	Subject           certificates.CSRSubject
	CertService       certificates.Service
}

type MiddlewaresList struct {
	AppCSRInfoMiddlewares []mux.MiddlewareFunc
	AppSigningMiddlewares []mux.MiddlewareFunc
	AppRenewalMiddlewares []mux.MiddlewareFunc

	RuntimeCSRInfoMiddlewares []mux.MiddlewareFunc
	RuntimeSigningMiddlewares []mux.MiddlewareFunc
	RuntimeRenewalMiddlewares []mux.MiddlewareFunc
}

type SignatureHandler interface {
	SignCSR(w http.ResponseWriter, r *http.Request)
}

type CSRGetInfoHandler interface {
	GetCSRInfo(w http.ResponseWriter, r *http.Request)
}

type ManagementGetInfoHandler interface {
	GetManagementInfo(w http.ResponseWriter, r *http.Request)
}

const apiSpecPath = "connectorapi.yaml"

func NewHandler(appHandlerCfg, runtimeHandlerCfg, appMngmtInfoHandlerCfg, runtimeMngmtInfoHandlerCfg Config, middlewaresList MiddlewaresList, globalMiddlewares []mux.MiddlewareFunc) http.Handler {
	router := mux.NewRouter()

	httphelpers.WithMiddlewares(globalMiddlewares, router)

	router.Path("/v1").Handler(http.RedirectHandler("/v1/api.yaml", http.StatusMovedPermanently)).Methods(http.MethodGet)
	router.Path("/v1/api.yaml").Handler(NewStaticFileHandler(apiSpecPath)).Methods(http.MethodGet)

	applicationInfoHandler := NewCSRInfoHandler(appHandlerCfg.TokenCreator, appHandlerCfg.ContextExtractor, appHandlerCfg.ManagementInfoURL, appHandlerCfg.Subject, appHandlerCfg.BaseURL)
	applicationManagementInfoHandler := NewManagementInfoHandler(appMngmtInfoHandlerCfg.ContextExtractor)
	applicationSignatureHandler := NewSignatureHandler(appHandlerCfg.CertService, appHandlerCfg.ContextExtractor)
	applicationRenewalHandler := NewSignatureHandler(appHandlerCfg.CertService, appHandlerCfg.ContextExtractor)

	csrApplicationRouter := router.PathPrefix("/v1/applications/signingRequests").Subrouter()
	csrApplicationRouter.HandleFunc("/info", applicationInfoHandler.GetCSRInfo).Methods(http.MethodGet)

	certApplicationRouter := router.PathPrefix("/v1/applications/certificates").Subrouter()
	certApplicationRouter.HandleFunc("", applicationSignatureHandler.SignCSR).Methods(http.MethodPost)

	httphelpers.WithMiddlewares(appHandlerCfg.Middlewares, csrApplicationRouter, certApplicationRouter)

	mngmtApplicationRouter := router.PathPrefix("/v1/applications/management").Subrouter()
	mngmtApplicationRouter.HandleFunc("/info", applicationManagementInfoHandler.GetManagementInfo).Methods(http.MethodGet)

	httphelpers.WithMiddlewares(appMngmtInfoHandlerCfg.Middlewares, mngmtApplicationRouter)

	runtimeInfoHandler := NewCSRInfoHandler(runtimeHandlerCfg.TokenCreator, runtimeHandlerCfg.ContextExtractor, runtimeHandlerCfg.ManagementInfoURL, runtimeHandlerCfg.Subject, runtimeHandlerCfg.BaseURL)
	runtimeManagementInfoHandler := NewManagementInfoHandler(runtimeMngmtInfoHandlerCfg.ContextExtractor)
	runtimeSignatureHandler := NewSignatureHandler(runtimeHandlerCfg.CertService, runtimeHandlerCfg.ContextExtractor)
	runtimeRenewalHandler := NewSignatureHandler(runtimeHandlerCfg.CertService, runtimeHandlerCfg.ContextExtractor)

	appRenewalRouter := router.Path("/v1/applications/certificates/renewals").Subrouter()
	// TODO - use middlewares
	appRenewalRouter.HandleFunc("", applicationRenewalHandler.SignCSR)


	csrRuntimesRouter := router.PathPrefix("/v1/runtimes/signingRequests").Subrouter()
	csrRuntimesRouter.HandleFunc("/info", runtimeInfoHandler.GetCSRInfo).Methods(http.MethodGet)

	certRuntimesRouter := router.PathPrefix("/v1/runtimes/certificates").Subrouter()
	certRuntimesRouter.HandleFunc("", runtimeSignatureHandler.SignCSR).Methods(http.MethodPost)

	httphelpers.WithMiddlewares(runtimeHandlerCfg.Middlewares, csrRuntimesRouter, certRuntimesRouter)

	mngmtRuntimeRouter := router.PathPrefix("/v1/runtimes/management").Subrouter()
	mngmtRuntimeRouter.HandleFunc("/info", runtimeManagementInfoHandler.GetManagementInfo).Methods(http.MethodGet)

	runtimeRenewalRouter := router.Path("/v1/runtimes/certificates/renewals").Subrouter()
	// TODO - use middlewares
	runtimeRenewalRouter.HandleFunc("", runtimeRenewalHandler.SignCSR)

	router.NotFoundHandler = errorhandler.NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = errorhandler.NewErrorHandler(405, "Method not allowed.")

	return router
}

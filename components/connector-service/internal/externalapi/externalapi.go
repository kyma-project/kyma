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
	TokenCreator                tokens.Creator
	ContextExtractor            clientcontext.ConnectorClientExtractor
	ManagementInfoURL           string
	ConnectorServiceBaseURL     string
	CertificateProtectedBaseURL string
	Subject                     certificates.CSRSubject
	CertService                 certificates.Service
}

type FunctionalMiddlewares struct {
	AppTokenResolverMiddleware      mux.MiddlewareFunc
	RuntimeTokenResolverMiddleware  mux.MiddlewareFunc
	RuntimeURLsMiddleware           mux.MiddlewareFunc
	AppContextFromSubjectMiddleware mux.MiddlewareFunc
}

type SignatureHandler interface {
	SignCSR(w http.ResponseWriter, r *http.Request)
}

type CSRInfoHandler interface {
	GetCSRInfo(w http.ResponseWriter, r *http.Request)
}

type ManagementInfoHandler interface {
	GetManagementInfo(w http.ResponseWriter, r *http.Request)
}

const apiSpecPath = "connectorapi.yaml"

type handlerBuilder struct {
	router         *mux.Router
	funcMiddlwares FunctionalMiddlewares
}

func NewHandlerBuilder(funcMiddlwares FunctionalMiddlewares, globalMiddlewares []mux.MiddlewareFunc) *handlerBuilder {
	router := mux.NewRouter()

	httphelpers.WithMiddlewares(router, globalMiddlewares...)

	router.Path("/v1").Handler(http.RedirectHandler("/v1/api.yaml", http.StatusMovedPermanently)).Methods(http.MethodGet)
	router.Path("/v1/api.yaml").Handler(NewStaticFileHandler(apiSpecPath)).Methods(http.MethodGet)

	router.NotFoundHandler = errorhandler.NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = errorhandler.NewErrorHandler(405, "Method not allowed.")

	return &handlerBuilder{
		router:         router,
		funcMiddlwares: funcMiddlwares,
	}
}

func (hb *handlerBuilder) WithApps(appHandlerCfg Config) {
	applicationInfoHandler := NewCSRInfoHandler(appHandlerCfg.TokenCreator, appHandlerCfg.ContextExtractor, appHandlerCfg.ManagementInfoURL, appHandlerCfg.Subject, appHandlerCfg.ConnectorServiceBaseURL)
	applicationRenewalHandler := NewSignatureHandler(appHandlerCfg.CertService, appHandlerCfg.ContextExtractor)
	applicationSignatureHandler := NewSignatureHandler(appHandlerCfg.CertService, appHandlerCfg.ContextExtractor)
	applicationManagementInfoHandler := NewManagementInfoHandler(appHandlerCfg.ContextExtractor, appHandlerCfg.CertificateProtectedBaseURL)

	csrApplicationRouter := hb.router.PathPrefix("/v1/applications/signingRequests").Subrouter()
	csrApplicationRouter.HandleFunc("/info", applicationInfoHandler.GetCSRInfo).Methods(http.MethodGet)
	httphelpers.WithMiddlewares(csrApplicationRouter, hb.funcMiddlwares.AppTokenResolverMiddleware, hb.funcMiddlwares.RuntimeURLsMiddleware)

	appRenewalRouter := hb.router.Path("/v1/applications/certificates/renewals").Subrouter()
	appRenewalRouter.HandleFunc("", applicationRenewalHandler.SignCSR).Methods(http.MethodPost)
	httphelpers.WithMiddlewares(appRenewalRouter, hb.funcMiddlwares.AppContextFromSubjectMiddleware)

	certApplicationRouter := hb.router.PathPrefix("/v1/applications/certificates").Subrouter()
	certApplicationRouter.HandleFunc("", applicationSignatureHandler.SignCSR).Methods(http.MethodPost)
	httphelpers.WithMiddlewares(certApplicationRouter, hb.funcMiddlwares.AppTokenResolverMiddleware)

	mngmtApplicationRouter := hb.router.PathPrefix("/v1/applications/management").Subrouter()
	mngmtApplicationRouter.HandleFunc("/info", applicationManagementInfoHandler.GetManagementInfo).Methods(http.MethodGet)
	httphelpers.WithMiddlewares(mngmtApplicationRouter, hb.funcMiddlwares.RuntimeURLsMiddleware, hb.funcMiddlwares.AppContextFromSubjectMiddleware)
}

func (hb *handlerBuilder) WithRuntimes(runtimeHandlerCfg Config) {
	runtimeInfoHandler := NewCSRInfoHandler(runtimeHandlerCfg.TokenCreator, runtimeHandlerCfg.ContextExtractor, runtimeHandlerCfg.ManagementInfoURL, runtimeHandlerCfg.Subject, runtimeHandlerCfg.ConnectorServiceBaseURL)
	runtimeRenewalHandler := NewSignatureHandler(runtimeHandlerCfg.CertService, runtimeHandlerCfg.ContextExtractor)
	runtimeSignatureHandler := NewSignatureHandler(runtimeHandlerCfg.CertService, runtimeHandlerCfg.ContextExtractor)
	runtimeManagementInfoHandler := NewManagementInfoHandler(runtimeHandlerCfg.ContextExtractor, runtimeHandlerCfg.CertificateProtectedBaseURL)

	csrRuntimesRouter := hb.router.PathPrefix("/v1/runtimes/signingRequests").Subrouter()
	csrRuntimesRouter.HandleFunc("/info", runtimeInfoHandler.GetCSRInfo).Methods(http.MethodGet)
	httphelpers.WithMiddlewares(csrRuntimesRouter, hb.funcMiddlwares.RuntimeTokenResolverMiddleware)

	runtimeRenewalRouter := hb.router.Path("/v1/runtimes/certificates/renewals").Subrouter()
	runtimeRenewalRouter.HandleFunc("", runtimeRenewalHandler.SignCSR).Methods(http.MethodPost)
	httphelpers.WithMiddlewares(runtimeRenewalRouter, hb.funcMiddlwares.AppContextFromSubjectMiddleware)

	certRuntimesRouter := hb.router.PathPrefix("/v1/runtimes/certificates").Subrouter()
	certRuntimesRouter.HandleFunc("", runtimeSignatureHandler.SignCSR).Methods(http.MethodPost)
	httphelpers.WithMiddlewares(certRuntimesRouter, hb.funcMiddlwares.RuntimeTokenResolverMiddleware)

	mngmtRuntimeRouter := hb.router.PathPrefix("/v1/runtimes/management").Subrouter()
	mngmtRuntimeRouter.HandleFunc("/info", runtimeManagementInfoHandler.GetManagementInfo).Methods(http.MethodGet)
	httphelpers.WithMiddlewares(mngmtRuntimeRouter, hb.funcMiddlwares.AppContextFromSubjectMiddleware)
}

func (hb *handlerBuilder) GetHandler() http.Handler {
	return hb.router
}

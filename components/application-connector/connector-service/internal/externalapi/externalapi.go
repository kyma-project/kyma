package externalapi

import (
	"net/http"

	loggingMiddlewares "github.com/kyma-project/kyma/components/connector-service/internal/logging/middlewares"
	"github.com/kyma-project/kyma/components/connector-service/internal/revocation"

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
	CertService                 certificates.Service
	RevokedCertsRepo            revocation.RevocationListRepository
	HeaderParser                certificates.HeaderParser
}

type FunctionalMiddlewares struct {
	AppTokenResolverMiddleware      mux.MiddlewareFunc
	RuntimeTokenResolverMiddleware  mux.MiddlewareFunc
	RuntimeURLsMiddleware           mux.MiddlewareFunc
	AppContextFromSubjectMiddleware mux.MiddlewareFunc
	CheckForRevokedCertMiddleware   mux.MiddlewareFunc
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
	applicationInfoHandler := NewCSRInfoHandler(appHandlerCfg.TokenCreator, appHandlerCfg.ContextExtractor, appHandlerCfg.ManagementInfoURL, appHandlerCfg.ConnectorServiceBaseURL)
	applicationRenewalHandler := NewSignatureHandler(appHandlerCfg.CertService, appHandlerCfg.ContextExtractor)
	applicationSignatureHandler := NewSignatureHandler(appHandlerCfg.CertService, appHandlerCfg.ContextExtractor)
	applicationManagementInfoHandler := NewManagementInfoHandler(appHandlerCfg.ContextExtractor, appHandlerCfg.CertificateProtectedBaseURL)
	applicationRevocationHandler := NewRevocationHandler(appHandlerCfg.RevokedCertsRepo, appHandlerCfg.HeaderParser)

	csrApplicationRouter := hb.router.PathPrefix("/v1/applications/signingRequests").Subrouter()
	csrApplicationRouter.HandleFunc("/info", applicationInfoHandler.GetCSRInfo).Methods(http.MethodGet)
	httphelpers.WithMiddlewares(
		csrApplicationRouter,
		hb.funcMiddlwares.AppTokenResolverMiddleware,
		hb.funcMiddlwares.RuntimeURLsMiddleware)

	appRenewalRouter := hb.router.Path("/v1/applications/certificates/renewals").Subrouter()
	appRenewalRouter.HandleFunc("", applicationRenewalHandler.SignCSR).Methods(http.MethodPost)
	renewalAuditLoggingMiddleware := hb.createRenewalAuditLogMiddleware(appHandlerCfg.ContextExtractor)
	httphelpers.WithMiddlewares(
		appRenewalRouter,
		hb.funcMiddlwares.AppContextFromSubjectMiddleware,
		renewalAuditLoggingMiddleware,
		hb.funcMiddlwares.CheckForRevokedCertMiddleware)

	appRevocationRouter := hb.router.Path("/v1/applications/certificates/revocations").Subrouter()
	appRevocationRouter.HandleFunc("", applicationRevocationHandler.Revoke).Methods(http.MethodPost)
	revocationAuditLoggingMiddleware := hb.createCertificateRevocationAuditLogMiddleware(appHandlerCfg.ContextExtractor)
	httphelpers.WithMiddlewares(
		appRevocationRouter,
		hb.funcMiddlwares.AppContextFromSubjectMiddleware,
		revocationAuditLoggingMiddleware)

	certApplicationRouter := hb.router.PathPrefix("/v1/applications/certificates").Subrouter()
	signingAuditLoggingMiddleware := hb.createCertificateGenerationAuditLogMiddleware(appHandlerCfg.ContextExtractor)
	certApplicationRouter.HandleFunc("", applicationSignatureHandler.SignCSR).Methods(http.MethodPost)
	httphelpers.WithMiddlewares(
		certApplicationRouter,
		hb.funcMiddlwares.AppTokenResolverMiddleware,
		signingAuditLoggingMiddleware)

	mngmtApplicationRouter := hb.router.PathPrefix("/v1/applications/management").Subrouter()
	mngmtApplicationRouter.HandleFunc("/info", applicationManagementInfoHandler.GetManagementInfo).Methods(http.MethodGet)
	httphelpers.WithMiddlewares(
		mngmtApplicationRouter,
		hb.funcMiddlwares.AppContextFromSubjectMiddleware,
		hb.funcMiddlwares.RuntimeURLsMiddleware)
}

func (hb *handlerBuilder) createRenewalAuditLogMiddleware(contextExtractor clientcontext.ConnectorClientExtractor) mux.MiddlewareFunc {
	return loggingMiddlewares.NewAuditLoggingMiddleware(contextExtractor, loggingMiddlewares.AuditLogMessages{
		StartingOperationMsg:   "Starting certificate renewal.",
		OperationSuccessfulMsg: "Certificate renewed successfully.",
		OperationFailedMsg:     "Certificate renewal failed.",
	}).Middleware
}

func (hb *handlerBuilder) createCertificateGenerationAuditLogMiddleware(contextExtractor clientcontext.ConnectorClientExtractor) mux.MiddlewareFunc {
	return loggingMiddlewares.NewAuditLoggingMiddleware(contextExtractor, loggingMiddlewares.AuditLogMessages{
		StartingOperationMsg:   "Starting certificate generation.",
		OperationSuccessfulMsg: "Certificate generated successfully.",
		OperationFailedMsg:     "Certificate generation failed.",
	}).Middleware
}

func (hb *handlerBuilder) createCertificateRevocationAuditLogMiddleware(contextExtractor clientcontext.ConnectorClientExtractor) mux.MiddlewareFunc {
	return loggingMiddlewares.NewAuditLoggingMiddleware(contextExtractor, loggingMiddlewares.AuditLogMessages{
		StartingOperationMsg:   "Starting certificate revocation.",
		OperationSuccessfulMsg: "Certificate revoked successfully.",
		OperationFailedMsg:     "Certificate revocation failed.",
	}).Middleware
}

func (hb *handlerBuilder) GetHandler() http.Handler {
	return hb.router
}

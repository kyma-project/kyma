package internalapi

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/revocation"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

type Config struct {
	TokenManager            tokens.Creator
	CSRInfoURL              string
	ContextExtractor        clientcontext.ConnectorClientExtractor
	RevokedCertsRepo        revocation.RevocationListRepository
	RevokedRuntimeCertsRepo revocation.RevocationListRepository
}

type FunctionalMiddlewares struct {
	ApplicationCtxMiddleware mux.MiddlewareFunc
	RuntimeCtxMiddleware     mux.MiddlewareFunc
}

type TokenHandler interface {
	CreateToken(w http.ResponseWriter, r *http.Request)
}

type handlerBuilder struct {
	router                *mux.Router
	functionalMiddlewares FunctionalMiddlewares
}

func NewHandlerBuilder(functionalMiddlewares FunctionalMiddlewares, globalMiddlewares []mux.MiddlewareFunc) *handlerBuilder {
	router := mux.NewRouter()
	httphelpers.WithMiddlewares(router, globalMiddlewares...)

	router.NotFoundHandler = errorhandler.NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = errorhandler.NewErrorHandler(405, "Method not allowed.")

	return &handlerBuilder{
		router:                router,
		functionalMiddlewares: functionalMiddlewares,
	}
}

func (hb *handlerBuilder) WithApps(appCfg Config) {
	appTokenHandler := NewTokenHandler(appCfg.TokenManager, appCfg.CSRInfoURL, appCfg.ContextExtractor)
	appRevocationHandler := NewRevocationHandler(appCfg.RevokedCertsRepo)

	applicationTokenRouter := hb.router.PathPrefix("/v1/applications").Subrouter()
	httphelpers.WithMiddlewares(applicationTokenRouter, hb.functionalMiddlewares.ApplicationCtxMiddleware)
	applicationTokenRouter.HandleFunc("/tokens", appTokenHandler.CreateToken).Methods(http.MethodPost)

	applicationRevocationRouter := hb.router.Path("/v1/applications/certificates/revocations").Subrouter()
	applicationRevocationRouter.HandleFunc("", appRevocationHandler.Revoke).Methods(http.MethodPost)
}

func (hb *handlerBuilder) WithRuntimes(runtimeCfg Config) {
	runtimeTokenHandler := NewTokenHandler(runtimeCfg.TokenManager, runtimeCfg.CSRInfoURL, runtimeCfg.ContextExtractor)
	runtimeRevocationHandler := NewRevocationHandler(runtimeCfg.RevokedRuntimeCertsRepo)

	clusterTokenRouter := hb.router.PathPrefix("/v1/runtimes").Subrouter()
	httphelpers.WithMiddlewares(clusterTokenRouter, hb.functionalMiddlewares.RuntimeCtxMiddleware)
	clusterTokenRouter.HandleFunc("/tokens", runtimeTokenHandler.CreateToken).Methods(http.MethodPost)

	runtimeRevocationRouter := hb.router.Path("/v1/runtimes/certificates/revocations").Subrouter()
	runtimeRevocationRouter.HandleFunc("", runtimeRevocationHandler.Revoke).Methods(http.MethodPost)
}

func (hb *handlerBuilder) GetHandler() http.Handler {
	return hb.router
}

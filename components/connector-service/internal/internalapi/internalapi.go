package internalapi

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

type Config struct {
	Middlewares      []mux.MiddlewareFunc
	TokenManager     tokens.Creator
	CSRInfoURL       string
	ContextExtractor clientcontext.ConnectorClientExtractor
}

type TokenHandler interface {
	CreateToken(w http.ResponseWriter, r *http.Request)
}

type handlerBuilder struct {
	router *mux.Router
}

func NewHandlerBuilder(globalMiddlewares []mux.MiddlewareFunc) *handlerBuilder {
	router := mux.NewRouter()
	httphelpers.WithMiddlewares(router, globalMiddlewares...)

	router.NotFoundHandler = errorhandler.NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = errorhandler.NewErrorHandler(405, "Method not allowed.")

	return &handlerBuilder{
		router: router,
	}
}

func (hb *handlerBuilder) WithApps(appCfg Config) {
	appTokenHandler := NewTokenHandler(appCfg.TokenManager, appCfg.CSRInfoURL, appCfg.ContextExtractor)

	applicationTokenRouter := hb.router.PathPrefix("/v1/applications").Subrouter()
	httphelpers.WithMiddlewares(applicationTokenRouter, appCfg.Middlewares...)
	applicationTokenRouter.HandleFunc("/tokens", appTokenHandler.CreateToken).Methods(http.MethodPost)
}

func (hb *handlerBuilder) WithRuntimes(runtimeCfg Config) {
	runtimeTokenHandler := NewTokenHandler(runtimeCfg.TokenManager, runtimeCfg.CSRInfoURL, runtimeCfg.ContextExtractor)

	clusterTokenRouter := hb.router.PathPrefix("/v1/runtimes").Subrouter()
	httphelpers.WithMiddlewares(clusterTokenRouter, runtimeCfg.Middlewares...)
	clusterTokenRouter.HandleFunc("/tokens", runtimeTokenHandler.CreateToken).Methods(http.MethodPost)
}

func (hb *handlerBuilder) GetHandler() http.Handler {
	return hb.router
}

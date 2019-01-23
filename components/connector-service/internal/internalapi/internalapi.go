package internalapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

type Config struct {
	Middlewares  []mux.MiddlewareFunc
	TokenService tokens.Service
	CSRInfoURL   string
	ParamsParser tokens.TokenParamsParser
}

type TokenHandler interface {
	CreateToken(w http.ResponseWriter, r *http.Request)
}

func NewHandler(globalMiddlewares []mux.MiddlewareFunc, appCfg Config, runtimeCfg Config) http.Handler {
	router := mux.NewRouter()
	httphelpers.WithMiddlewares(router, globalMiddlewares)

	appTokenHandler := NewTokenHandler(appCfg.TokenService, appCfg.CSRInfoURL, appCfg.ParamsParser)

	applicationTokenRouter := router.PathPrefix("/v1/applications").Subrouter()
	httphelpers.WithMiddlewares(applicationTokenRouter, appCfg.Middlewares)
	applicationTokenRouter.HandleFunc("/tokens", appTokenHandler.CreateToken).Methods(http.MethodPost)

	runtimeTokenHandler := NewTokenHandler(runtimeCfg.TokenService, runtimeCfg.CSRInfoURL, runtimeCfg.ParamsParser)

	clusterTokenRouter := router.PathPrefix("/v1/runtimes").Subrouter()
	httphelpers.WithMiddlewares(clusterTokenRouter, runtimeCfg.Middlewares)
	clusterTokenRouter.HandleFunc("/tokens", runtimeTokenHandler.CreateToken).Methods(http.MethodPost)

	router.NotFoundHandler = errorhandler.NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = errorhandler.NewErrorHandler(405, "Method not allowed.")

	return router
}

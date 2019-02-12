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

func NewHandler(globalMiddlewares []mux.MiddlewareFunc, appCfg Config, runtimeCfg Config) http.Handler {
	router := mux.NewRouter()
	httphelpers.WithMiddlewares(globalMiddlewares, router)

	appTokenHandler := NewTokenHandler(appCfg.TokenManager, appCfg.CSRInfoURL, appCfg.ContextExtractor)

	applicationTokenRouter := router.PathPrefix("/v1/applications").Subrouter()
	httphelpers.WithMiddlewares(appCfg.Middlewares, applicationTokenRouter)
	applicationTokenRouter.HandleFunc("/tokens", appTokenHandler.CreateToken).Methods(http.MethodPost)

	runtimeTokenHandler := NewTokenHandler(runtimeCfg.TokenManager, runtimeCfg.CSRInfoURL, runtimeCfg.ContextExtractor)

	clusterTokenRouter := router.PathPrefix("/v1/runtimes").Subrouter()
	httphelpers.WithMiddlewares(runtimeCfg.Middlewares, clusterTokenRouter)
	clusterTokenRouter.HandleFunc("/tokens", runtimeTokenHandler.CreateToken).Methods(http.MethodPost)

	router.NotFoundHandler = errorhandler.NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = errorhandler.NewErrorHandler(405, "Method not allowed.")

	return router
}

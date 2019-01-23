package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/kyma-project/kyma/components/connector-service/internal/httpcontext"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/internalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/middlewares"
	"github.com/kyma-project/kyma/components/connector-service/internal/monitoring"
	"github.com/kyma-project/kyma/components/connector-service/internal/secrets"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	appCSRInfoFmt     = "https://%s/v1/applications/csr/info"
	runtimeCSRInfoFmt = "https://%s/v1/runtimes/csr/info"
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting Certificate Service.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	env := parseEnv()
	log.Infof("Environment variables: %s", env)

	tokenCache := tokencache.NewTokenCache(options.tokenExpirationMinutes)
	tokenGenerator := tokens.NewTokenGenerator(options.tokenLength)
	tokenService := tokens.NewTokenService(tokenCache, tokenGenerator.NewToken)
	certUtil := certificates.NewCertificateUtility()

	middlewares, appErr := monitoring.SetupMonitoringMiddleware()
	if appErr != nil {
		log.Errorf("Error while setting up monitoring: %s", appErr)
	}

	externalHandler := newExternalHandler(tokenCache, certUtil, tokenGenerator, options, env, middlewares)
	internalHandler := newInternalHandler(tokenService, options, middlewares)

	externalSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(options.externalAPIPort),
		Handler: externalHandler,
	}

	internalSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(options.internalAPIPort),
		Handler: internalHandler,
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		log.Info(externalSrv.ListenAndServe())
	}()

	go func() {
		log.Info(internalSrv.ListenAndServe())
	}()

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Info(http.ListenAndServe(":9090", nil))
	}()

	wg.Wait()
}

func newExternalHandler(cache tokencache.TokenCache, utility certificates.CertificateUtility, tokenGenerator tokens.TokenGenerator, opts *options, env *environment, middlewares []mux.MiddlewareFunc) http.Handler {
	secretsRepository, appErr := newSecretsRepository(opts.namespace)
	if appErr != nil {
		log.Infof("Failed to create secrets repository. %s", appErr.Error())
		return errorhandler.NewErrorHandler(500, fmt.Sprintf("Failed to create secrets repository: %s", appErr.Error()))
	}

	subjectValues := certificates.CSRSubject{
		Country:            env.country,
		Organization:       env.organization,
		OrganizationalUnit: env.organizationalUnit,
		Locality:           env.locality,
		Province:           env.province,
	}
	rh := externalapi.NewSignatureHandler(cache, utility, secretsRepository, opts.connectorServiceHost, opts.domainName, subjectValues)
	ih := externalapi.NewInfoHandler(cache, tokenGenerator, opts.connectorServiceHost, opts.domainName, subjectValues)
	return externalapi.NewHandler(rh, ih, middlewares)
}

func newInternalHandler(tokenService tokens.Service, opts *options, globalMiddlewares []mux.MiddlewareFunc) http.Handler {

	applicationCtxMiddleware := middlewares.NewApplicationContextMiddleware()
	clusterCtxMiddleware := middlewares.NewClusterContextMiddleware(opts.tenant, opts.group)

	appHandlerMiddlewares := []mux.MiddlewareFunc{applicationCtxMiddleware.Middleware}
	appHandlerConfig := internalapi.Config{
		Middlewares:      appHandlerMiddlewares,
		TokenCreator:     tokenService,
		CSRInfoURL:       fmt.Sprintf(appCSRInfoFmt, opts.connectorServiceHost),
		ContextExtractor: httpcontext.ExtractSerializableApplicationContext,
	}

	runtimeHandlerMiddlewares := []mux.MiddlewareFunc{clusterCtxMiddleware.Middleware}
	runtimeHandlerConfig := internalapi.Config{
		Middlewares:      runtimeHandlerMiddlewares,
		TokenCreator:     tokenService,
		CSRInfoURL:       fmt.Sprintf(runtimeCSRInfoFmt, opts.connectorServiceHost),
		ContextExtractor: httpcontext.ExtractSerializableClusterContext,
	}

	return internalapi.NewHandler(globalMiddlewares, appHandlerConfig, runtimeHandlerConfig)
}

func newSecretsRepository(namespace string) (secrets.Repository, apperrors.AppError) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, apperrors.Internal("failed to read k8s in-cluster configuration, %s", err)
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s core client, %s", err)
	}

	sei := coreClientset.CoreV1().Secrets(namespace)

	return secrets.NewRepository(sei), nil
}

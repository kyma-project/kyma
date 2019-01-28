package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/kyma-project/kyma/components/connector-service/internal/logging"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	clientcontextmiddlewares "github.com/kyma-project/kyma/components/connector-service/internal/clientcontext/middlewares"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi/middlewares"
	"github.com/kyma-project/kyma/components/connector-service/internal/internalapi"
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
	appCSRInfoFmt            = "https://%s/v1/applications/signingRequests/info"
	runtimeCSRInfoFmt        = "https://%s/v1/runtimes/signingRequests/info"
	appCertificateURLFmt     = "https://%s/v1/applications/certificates"
	runtimeCertificateURLFmt = "https://%s/v1/runtimes/certificates"
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

	globalMiddlewares, appErr := monitoring.SetupMonitoringMiddleware()
	if appErr != nil {
		log.Errorf("Error while setting up monitoring: %s", appErr)
	}

	if options.requestLogging {
		globalMiddlewares = append(globalMiddlewares, logging.NewLoggingMiddleware().Middleware)
	}

	internalHandler := newInternalHandler(tokenService, options, globalMiddlewares)
	externalHandler := newExternalHandler(tokenService, options, env, globalMiddlewares)

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

func newExternalHandler(tokenService tokens.Service, opts *options, env *environment, globalMiddlewares []mux.MiddlewareFunc) http.Handler {
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

	certificateService := certificates.NewCertificateService(secretsRepository, certificates.NewCertificateUtility(), opts.caSecretName, subjectValues)

	appTokenResolverMiddleware := middlewares.NewTokenResolverMiddleware(tokenService, clientcontext.ResolveApplicationContextExtender)
	appAPIUrlsGenerator := externalapi.NewApplicationApiUrlsStrategy(opts.appRegistryHost, opts.eventsHost, opts.getInfoURL, opts.connectorServiceHost)

	appHandlerConfig := externalapi.Config{
		TokenService:     tokenService,
		CertificateURL:   fmt.Sprintf(appCertificateURLFmt, opts.connectorServiceHost),
		Subject:          subjectValues,
		Middlewares:      []mux.MiddlewareFunc{appTokenResolverMiddleware.Middleware},
		ContextExtractor: clientcontext.ExtractApplicationContext,
		APIUrlsGenerator: appAPIUrlsGenerator,
		CertService:      certificateService,
	}

	clusterTokenResolverMiddleware := middlewares.NewTokenResolverMiddleware(tokenService, clientcontext.ResolveClusterContextExtender)
	runtimeAPIUrlsGenerator := externalapi.NewRuntimeApiUrlsStrategy(opts.connectorServiceHost)

	runtimeHandlerConfig := externalapi.Config{
		TokenService:     tokenService,
		CertificateURL:   fmt.Sprintf(runtimeCertificateURLFmt, opts.connectorServiceHost),
		Subject:          subjectValues,
		Middlewares:      []mux.MiddlewareFunc{clusterTokenResolverMiddleware.Middleware},
		ContextExtractor: clientcontext.ExtractClusterContext,
		APIUrlsGenerator: runtimeAPIUrlsGenerator,
		CertService:      certificateService,
	}

	return externalapi.NewHandler(appHandlerConfig, runtimeHandlerConfig, globalMiddlewares)
}

func newInternalHandler(tokenService tokens.Service, opts *options, globalMiddlewares []mux.MiddlewareFunc) http.Handler {

	clusterCtxMiddleware := clientcontextmiddlewares.NewClusterContextMiddleware(opts.tenant, opts.group)
	applicationCtxMiddleware := clientcontextmiddlewares.NewApplicationContextMiddleware(clusterCtxMiddleware)

	appHandlerMiddlewares := []mux.MiddlewareFunc{applicationCtxMiddleware.Middleware}
	appHandlerConfig := internalapi.Config{
		Middlewares:      appHandlerMiddlewares,
		TokenCreator:     tokenService,
		CSRInfoURL:       fmt.Sprintf(appCSRInfoFmt, opts.connectorServiceHost),
		ContextExtractor: clientcontext.ExtractApplicationContext,
	}

	runtimeHandlerMiddlewares := []mux.MiddlewareFunc{clusterCtxMiddleware.Middleware}
	runtimeHandlerConfig := internalapi.Config{
		Middlewares:      runtimeHandlerMiddlewares,
		TokenCreator:     tokenService,
		CSRInfoURL:       fmt.Sprintf(runtimeCSRInfoFmt, opts.connectorServiceHost),
		ContextExtractor: clientcontext.ExtractClusterContext,
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

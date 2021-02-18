package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	clientcontextmiddlewares "github.com/kyma-project/kyma/components/connector-service/internal/clientcontext/middlewares"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi/middlewares"
	"github.com/kyma-project/kyma/components/connector-service/internal/internalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/revocation"
	certificateMiddlewares "github.com/kyma-project/kyma/components/connector-service/internal/revocation/middlewares"
	"github.com/kyma-project/kyma/components/connector-service/internal/secrets"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	appCSRInfoFmt     = "https://%s/v1/applications/signingRequests/info"
	runtimeCSRInfoFmt = "https://%s/v1/runtimes/signingRequests/info"
	AppURLFormat      = "https://%s/v1/applications"
	RuntimeURLFormat  = "https://%s/v1/runtimes"
)

type Handlers struct {
	internalAPI http.Handler
	externalAPI http.Handler
}

func createAPIHandlers(tokenManager tokens.Manager, tokenCreatorProvider tokens.TokenCreatorProvider, opts *options, env *environment, globalMiddlewares []mux.MiddlewareFunc) Handlers {

	coreClientSet, appErr := newCoreClientSet()
	if appErr != nil {
		log.Infof("Failed to initialize Kubernetes client. %s", appErr.Error())
		errorHandler := errorhandler.NewErrorHandler(500, fmt.Sprintf("Failed to initialize Kubernetes client: %s", appErr.Error()))
		return Handlers{
			internalAPI: errorHandler,
			externalAPI: errorHandler,
		}
	}

	revokedCertsRepo := newRevokedCertsRepository(coreClientSet, opts.namespace, opts.revocationConfigMapName)
	secretsRepository := newSecretsRepository(coreClientSet)

	subjectValues := certificates.CSRSubject{
		Country:            env.country,
		Organization:       env.organization,
		OrganizationalUnit: env.organizationalUnit,
		Locality:           env.locality,
		Province:           env.province,
	}

	contextExtractor := clientcontext.NewContextExtractor(subjectValues)

	return Handlers{
		internalAPI: newInternalHandler(tokenCreatorProvider, opts, globalMiddlewares, revokedCertsRepo, contextExtractor),
		externalAPI: newExternalHandler(tokenManager, tokenCreatorProvider, opts, env, globalMiddlewares, secretsRepository, revokedCertsRepo, contextExtractor),
	}
}

func newExternalHandler(tokenManager tokens.Manager, tokenCreatorProvider tokens.TokenCreatorProvider, opts *options, env *environment, globalMiddlewares []mux.MiddlewareFunc,
	secretsRepository secrets.Repository, revocationListRepository revocation.RevocationListRepository, contextExtractor *clientcontext.ContextExtractor) http.Handler {

	lookupEnabled := clientcontext.LookupEnabledType(opts.lookupEnabled)

	lookupService := middlewares.NewGraphQLLookupService()

	headerParser := certificates.NewHeaderParser(env.country, env.province, env.locality, env.organization, env.organizationalUnit, opts.central)

	appCertificateService := certificates.NewCertificateService(secretsRepository, certificates.NewCertificateUtility(opts.appCertificateValidityTime), opts.caSecretName, opts.rootCACertificateSecretName)

	appTokenResolverMiddleware := middlewares.NewTokenResolverMiddleware(tokenManager, clientcontext.NewApplicationContextExtender)
	clusterTokenResolverMiddleware := middlewares.NewTokenResolverMiddleware(tokenManager, clientcontext.NewClusterContextExtender)
	runtimeURLsMiddleware := middlewares.NewRuntimeURLsMiddleware(opts.gatewayBaseURL, opts.lookupConfigMapPath, lookupEnabled, clientcontext.ExtractApplicationContext, lookupService)
	contextFromSubjMiddleware := clientcontextmiddlewares.NewContextFromSubjMiddleware(headerParser, opts.central)
	checkForRevokedCertMiddleware := certificateMiddlewares.NewRevocationCheckMiddleware(revocationListRepository, headerParser)

	functionalMiddlewares := externalapi.FunctionalMiddlewares{
		AppTokenResolverMiddleware:      appTokenResolverMiddleware.Middleware,
		RuntimeTokenResolverMiddleware:  clusterTokenResolverMiddleware.Middleware,
		RuntimeURLsMiddleware:           runtimeURLsMiddleware.Middleware,
		AppContextFromSubjectMiddleware: contextFromSubjMiddleware.Middleware,
		CheckForRevokedCertMiddleware:   checkForRevokedCertMiddleware.Middleware,
	}

	handlerBuilder := externalapi.NewHandlerBuilder(functionalMiddlewares, globalMiddlewares)

	appTokenTTLMinutes := time.Duration(opts.appTokenExpirationMinutes) * time.Minute

	appHandlerConfig := externalapi.Config{
		TokenCreator:                tokenCreatorProvider.WithTTL(appTokenTTLMinutes),
		ManagementInfoURL:           opts.appsInfoURL,
		ConnectorServiceBaseURL:     fmt.Sprintf(AppURLFormat, opts.connectorServiceHost),
		CertificateProtectedBaseURL: fmt.Sprintf(AppURLFormat, opts.certificateProtectedHost),
		ContextExtractor:            contextExtractor.CreateApplicationClientContextService,
		CertService:                 appCertificateService,
		RevokedCertsRepo:            revocationListRepository,
		HeaderParser:                headerParser,
	}

	handlerBuilder.WithApps(appHandlerConfig)

	if opts.central {
		runtimeCertificateService := certificates.NewCertificateService(secretsRepository, certificates.NewCertificateUtility(opts.runtimeCertificateValidityTime), opts.caSecretName, opts.rootCACertificateSecretName)
		runtimeTokenTTLMinutes := time.Duration(opts.runtimeTokenExpirationMinutes) * time.Minute

		runtimeHandlerConfig := externalapi.Config{
			TokenCreator:                tokenCreatorProvider.WithTTL(runtimeTokenTTLMinutes),
			ManagementInfoURL:           opts.runtimesInfoURL,
			ConnectorServiceBaseURL:     fmt.Sprintf(RuntimeURLFormat, opts.connectorServiceHost),
			CertificateProtectedBaseURL: fmt.Sprintf(RuntimeURLFormat, opts.certificateProtectedHost),
			ContextExtractor:            contextExtractor.CreateClusterClientContextService,
			CertService:                 runtimeCertificateService,
			RevokedCertsRepo:            revocationListRepository,
			HeaderParser:                headerParser,
		}

		handlerBuilder.WithRuntimes(runtimeHandlerConfig)
	}

	return handlerBuilder.GetHandler()
}

func newInternalHandler(tokenManagerProvider tokens.TokenCreatorProvider, opts *options, globalMiddlewares []mux.MiddlewareFunc,
	revocationListRepository revocation.RevocationListRepository, contextExtractor *clientcontext.ContextExtractor) http.Handler {

	clusterCtxEnabled := clientcontext.CtxEnabledType(opts.central)
	clusterContextStrategy := clientcontext.NewClusterContextStrategy(clusterCtxEnabled)

	clusterCtxMiddleware := clientcontextmiddlewares.NewClusterContextMiddleware(clusterContextStrategy)
	applicationCtxMiddleware := clientcontextmiddlewares.NewApplicationContextMiddleware(clusterContextStrategy)

	appTokenTTLMinutes := time.Duration(opts.appTokenExpirationMinutes) * time.Minute
	appHandlerConfig := internalapi.Config{
		TokenManager:     tokenManagerProvider.WithTTL(appTokenTTLMinutes),
		CSRInfoURL:       fmt.Sprintf(appCSRInfoFmt, opts.connectorServiceHost),
		ContextExtractor: contextExtractor.CreateApplicationClientContextService,
		RevokedCertsRepo: revocationListRepository,
	}

	handlerBuilder := internalapi.NewHandlerBuilder(internalapi.FunctionalMiddlewares{
		ApplicationCtxMiddleware: applicationCtxMiddleware.Middleware,
		RuntimeCtxMiddleware:     clusterCtxMiddleware.Middleware,
	}, globalMiddlewares)

	handlerBuilder.WithApps(appHandlerConfig)

	if opts.central {
		runtimeTokenTTLMinutes := time.Duration(opts.runtimeTokenExpirationMinutes) * time.Minute
		runtimeHandlerConfig := internalapi.Config{
			TokenManager:            tokenManagerProvider.WithTTL(runtimeTokenTTLMinutes),
			CSRInfoURL:              fmt.Sprintf(runtimeCSRInfoFmt, opts.connectorServiceHost),
			ContextExtractor:        contextExtractor.CreateClusterClientContextService,
			RevokedRuntimeCertsRepo: revocationListRepository,
		}

		handlerBuilder.WithRuntimes(runtimeHandlerConfig)
	}

	return handlerBuilder.GetHandler()
}

func newSecretsRepository(coreClientSet *kubernetes.Clientset) secrets.Repository {
	sei := coreClientSet.CoreV1()

	return secrets.NewRepository(func(namespace string) secrets.Manager {
		return sei.Secrets(namespace)
	})
}

func newRevokedCertsRepository(coreClientSet *kubernetes.Clientset, namespace, revocationSecretName string) revocation.RevocationListRepository {
	cmi := coreClientSet.CoreV1().ConfigMaps(namespace)

	return revocation.NewRepository(cmi, revocationSecretName)
}

func newCoreClientSet() (*kubernetes.Clientset, apperrors.AppError) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, apperrors.Internal("failed to read k8s in-cluster configuration, %s", err)
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s core client, %s", err)
	}

	return coreClientset, nil
}

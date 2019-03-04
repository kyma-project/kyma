package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	certificateMiddlewares "github.com/kyma-project/kyma/components/connector-service/internal/certificates/middlewares"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates/revocationlist"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	clientcontextmiddlewares "github.com/kyma-project/kyma/components/connector-service/internal/clientcontext/middlewares"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi/middlewares"
	"github.com/kyma-project/kyma/components/connector-service/internal/internalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/secrets"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"net/http"
	"time"
)

const (
	appCSRInfoFmt     = "https://%s/v1/applications/signingRequests/info"
	runtimeCSRInfoFmt = "https://%s/v1/runtimes/signingRequests/info"
	AppURLFormat      = "https://%s/v1/applications"
	RuntimeURLFormat  = "https://%s/v1/runtimes"
)

func newExternalHandler(tokenManager tokens.Manager, tokenCreatorProvider tokens.TokenCreatorProvider,
	opts *options, env *environment, globalMiddlewares []mux.MiddlewareFunc, revokedAppCertsRepo, revokedRuntimeCertsRepo revocationlist.RevocationListRepository) http.Handler {

	headersRequired := clientcontext.HeadersRequiredType(opts.central)
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

	certificateService := certificates.NewCertificateService(secretsRepository, certificates.NewCertificateUtility(opts.certificateValidityTime), opts.caSecretName, subjectValues)

	appTokenResolverMiddleware := middlewares.NewTokenResolverMiddleware(tokenManager, clientcontext.NewApplicationContextExtender)
	clusterTokenResolverMiddleware := middlewares.NewTokenResolverMiddleware(tokenManager, clientcontext.NewClusterContextExtender)
	runtimeURLsMiddleware := middlewares.NewRuntimeURLsMiddleware(opts.gatewayHost, headersRequired)
	appContextFromSubjMiddleware := clientcontextmiddlewares.NewAppContextFromSubjMiddleware()
	checkForRevokedCertMiddleware := certificateMiddlewares.NewRevocationCheckMiddleware(revokedAppCertsRepo)

	functionalMiddlewares := externalapi.FunctionalMiddlewares{
		AppTokenResolverMiddleware:      appTokenResolverMiddleware.Middleware,
		RuntimeTokenResolverMiddleware:  clusterTokenResolverMiddleware.Middleware,
		RuntimeURLsMiddleware:           runtimeURLsMiddleware.Middleware,
		AppContextFromSubjectMiddleware: appContextFromSubjMiddleware.Middleware,
		CheckForRevokedCertMiddleware:   checkForRevokedCertMiddleware.Middleware,
	}

	handlerBuilder := externalapi.NewHandlerBuilder(functionalMiddlewares, globalMiddlewares)

	appTokenTTLMinutes := time.Duration(opts.appTokenExpirationMinutes) * time.Minute

	appHandlerConfig := externalapi.Config{
		TokenCreator:                tokenCreatorProvider.WithTTL(appTokenTTLMinutes),
		ManagementInfoURL:           opts.appsInfoURL,
		ConnectorServiceBaseURL:     fmt.Sprintf(AppURLFormat, opts.connectorServiceHost),
		CertificateProtectedBaseURL: fmt.Sprintf(AppURLFormat, opts.certificateProtectedHost),
		Subject:                     subjectValues,
		ContextExtractor:            clientcontext.CreateApplicationClientContextService,
		CertService:                 certificateService,
		RevokedApplicationCertsRepo: revokedAppCertsRepo,
	}

	handlerBuilder.WithApps(appHandlerConfig)

	if opts.central {
		runtimeTokenTTLMinutes := time.Duration(opts.runtimeTokenExpirationMinutes) * time.Minute

		runtimeHandlerConfig := externalapi.Config{
			TokenCreator:                tokenCreatorProvider.WithTTL(runtimeTokenTTLMinutes),
			ManagementInfoURL:           opts.runtimesInfoURL,
			ConnectorServiceBaseURL:     fmt.Sprintf(RuntimeURLFormat, opts.connectorServiceHost),
			CertificateProtectedBaseURL: fmt.Sprintf(RuntimeURLFormat, opts.certificateProtectedHost),
			Subject:                     subjectValues,
			ContextExtractor:            clientcontext.CreateClusterClientContextService,
			CertService:                 certificateService,
			RevokedRuntimeCertsRepo:     revokedRuntimeCertsRepo,
		}

		handlerBuilder.WithRuntimes(runtimeHandlerConfig)
	}

	return handlerBuilder.GetHandler()
}

func newInternalHandler(tokenManagerProvider tokens.TokenCreatorProvider, opts *options, globalMiddlewares []mux.MiddlewareFunc, revokedAppCertsRepo, revokedRuntimeCertsRepo revocationlist.RevocationListRepository) http.Handler {

	ctxRequired := clientcontext.CtxRequiredType(opts.central)

	clusterCtxMiddleware := clientcontextmiddlewares.NewClusterContextMiddleware(ctxRequired)
	applicationCtxMiddleware := clientcontextmiddlewares.NewApplicationContextMiddleware(clusterCtxMiddleware)

	appTokenTTLMinutes := time.Duration(opts.appTokenExpirationMinutes) * time.Minute
	appHandlerConfig := internalapi.Config{
		TokenManager:                tokenManagerProvider.WithTTL(appTokenTTLMinutes),
		CSRInfoURL:                  fmt.Sprintf(appCSRInfoFmt, opts.connectorServiceHost),
		ContextExtractor:            clientcontext.CreateApplicationClientContextService,
		RevokedApplicationCertsRepo: revokedAppCertsRepo,
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
			ContextExtractor:        clientcontext.CreateClusterClientContextService,
			RevokedRuntimeCertsRepo: revokedRuntimeCertsRepo,
		}

		handlerBuilder.WithRuntimes(runtimeHandlerConfig)
	}

	return handlerBuilder.GetHandler()
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

func newRevokedAppCertsRepo() revocationlist.RevocationListRepository {
	return revocationlist.NewRepository()
}

func newRevokedRuntimeCertsRepo() revocationlist.RevocationListRepository {
	return revocationlist.NewRepository()
}

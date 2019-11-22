package main

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/internal/csrf"
	csrfClient "github.com/kyma-project/kyma/components/application-gateway/internal/csrf/client"
	csrfStrategy "github.com/kyma-project/kyma/components/application-gateway/internal/csrf/strategy"
	"github.com/kyma-project/kyma/components/application-gateway/internal/externalapi"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/secrets"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/application-gateway/internal/proxy"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/httptools"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting Application Gateway.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	serviceDefinitionService, err := newServiceDefinitionService(
		options.namespace,
		options.application,
	)

	if err != nil {
		log.Errorf("Unable to create ServiceDefinitionService: '%s'", err.Error())
	}

	internalHandler := newInternalHandler(serviceDefinitionService, options)
	externalHandler := externalapi.NewHandler()

	if options.requestLogging {
		internalHandler = httptools.RequestLogger("Internal handler: ", internalHandler)
		externalHandler = httptools.RequestLogger("External handler: ", externalHandler)
	}

	externalSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(options.externalAPIPort),
		Handler:      externalHandler,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	internalSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(options.proxyPort),
		Handler:      internalHandler,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		log.Info(externalSrv.ListenAndServe())
	}()

	go func() {
		log.Info(internalSrv.ListenAndServe())
	}()

	wg.Wait()
}

func newInternalHandler(serviceDefinitionService metadata.ServiceDefinitionService, options *options) http.Handler {
	if serviceDefinitionService != nil {

		authStrategyFactory := newAuthenticationStrategyFactory(options.proxyTimeout)
		csrfCl := newCSRFClient(options.proxyTimeout)
		csrfTokenStrategyFactory := csrfStrategy.NewTokenStrategyFactory(csrfCl)

		proxyConfig := proxy.Config{
			SkipVerify:              options.skipVerify,
			ProxyTimeout:            options.proxyTimeout,
			Application:             options.application,
			ProxyCacheTTL:           options.proxyCacheTTL,
			AnnotatePassportHeaders: options.annotatePassportHeaders,
			RedisURL:                options.redisURL,
		}
		return proxy.New(serviceDefinitionService, authStrategyFactory, csrfTokenStrategyFactory, proxyConfig)
	}
	return proxy.NewInvalidStateHandler("Application Gateway is not initialized properly")
}

func newAuthenticationStrategyFactory(oauthClientTimeout int) authorization.StrategyFactory {
	return authorization.NewStrategyFactory(authorization.FactoryConfiguration{
		OAuthClientTimeout: oauthClientTimeout,
	})
}

func newServiceDefinitionService(namespace string, application string) (metadata.ServiceDefinitionService, apperrors.AppError) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, apperrors.Internal("failed to read k8s in-cluster configuration, %s", err)
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s core client, %s", err)
	}

	applicationServiceRepository, apperror := newApplicationRepository(k8sConfig, application)
	if apperror != nil {
		return nil, apperror
	}

	secretsRepository := newSecretsRepository(coreClientset, namespace, application)

	serviceAPIService := serviceapi.NewService(secretsRepository)

	return metadata.NewServiceDefinitionService(serviceAPIService, applicationServiceRepository), nil
}

func newApplicationRepository(config *restclient.Config, name string) (applications.ServiceRepository, apperrors.AppError) {
	applicationClientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s application client, %s", err)
	}

	rei := applicationClientset.ApplicationconnectorV1alpha1().Applications()

	return applications.NewServiceRepository(name, rei), nil
}

func newSecretsRepository(coreClientset *kubernetes.Clientset, namespace, application string) secrets.Repository {
	sei := coreClientset.CoreV1().Secrets(namespace)

	return secrets.NewRepository(sei, application)
}

func newCSRFClient(timeout int) csrf.Client {
	cache := csrfClient.NewTokenCache()
	client := &http.Client{}
	return csrfClient.New(timeout, cache, client)
}

package main

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	csrfClient "github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf/client"
	csrfStrategy "github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf/strategy"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/externalapi"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/secrets"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/proxy"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httptools"
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

	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.Fatalf("Error reading in cluster config: %s", err.Error())
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatalf("Error creating core clientset: %s", err.Error())
	}

	serviceDefinitionService, err := newServiceDefinitionService(
		k8sConfig,
		coreClientset,
		options.namespace,
	)
	if err != nil {
		log.Errorf("Unable to create ServiceDefinitionService: '%s'", err.Error())
		os.Exit(1)
	}

	internalHandler := newInternalHandler(serviceDefinitionService, options)
	internalHandlerForCompass := newInternalHandlerForCompass(serviceDefinitionService, options)
	externalHandler := externalapi.NewHandler()

	if options.requestLogging {
		internalHandler = httptools.RequestLogger("Internal handler: ", internalHandler)
		internalHandlerForCompass = httptools.RequestLogger("Internal handler: ", internalHandlerForCompass)
		externalHandler = httptools.RequestLogger("External handler: ", externalHandler)
	}

	externalSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(options.externalAPIPort),
		Handler:      externalHandler,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	internalSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(options.proxyOSPort),
		Handler:      internalHandler,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	internalSrvCompass := &http.Server{
		Addr:         ":" + strconv.Itoa(options.proxyMPSPort),
		Handler:      internalHandlerForCompass,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	// TODO: handle case when server fails
	wg := &sync.WaitGroup{}

	wg.Add(2)
	go func() {
		log.Fatal(externalSrv.ListenAndServe())
	}()

	if options.disableLegacyConnectivity {
		go func() {
			log.Fatal(internalSrvCompass.ListenAndServe())
		}()
	} else {
		go func() {
			log.Fatal(internalSrv.ListenAndServe())
		}()
	}

	wg.Wait()
}

func newInternalHandler(serviceDefinitionService metadata.ServiceDefinitionService, options *options) http.Handler {
	authStrategyFactory := newAuthenticationStrategyFactory(options.proxyTimeout)
	csrfCl := newCSRFClient(options.proxyTimeout)
	csrfTokenStrategyFactory := csrfStrategy.NewTokenStrategyFactory(csrfCl)

	return proxy.New(serviceDefinitionService, authStrategyFactory, csrfTokenStrategyFactory, getProxyConfig(options))
}

func newInternalHandlerForCompass(serviceDefinitionService metadata.ServiceDefinitionService, options *options) http.Handler {
	authStrategyFactory := newAuthenticationStrategyFactory(options.proxyTimeout)
	csrfCl := newCSRFClient(options.proxyTimeout)
	csrfTokenStrategyFactory := csrfStrategy.NewTokenStrategyFactory(csrfCl)

	return proxy.NewForCompass(serviceDefinitionService, authStrategyFactory, csrfTokenStrategyFactory, getProxyConfig(options))
}

func getProxyConfig(options *options) proxy.Config {
	return proxy.Config{
		SkipVerify:    options.skipVerify,
		ProxyTimeout:  options.proxyTimeout,
		ProxyCacheTTL: options.proxyCacheTTL,
	}
}

func newAuthenticationStrategyFactory(oauthClientTimeout int) authorization.StrategyFactory {
	return authorization.NewStrategyFactory(authorization.FactoryConfiguration{
		OAuthClientTimeout: oauthClientTimeout,
	})
}

func newServiceDefinitionService(k8sConfig *restclient.Config, coreClientset kubernetes.Interface, namespace string) (metadata.ServiceDefinitionService, error) {
	applicationServiceRepository, apperror := newApplicationRepository(k8sConfig)
	if apperror != nil {
		return nil, apperror
	}

	secretsRepository := newSecretsRepository(coreClientset, namespace)

	serviceAPIService := serviceapi.NewService(secretsRepository)

	return metadata.NewServiceDefinitionService(serviceAPIService, applicationServiceRepository), nil
}

func newApplicationRepository(config *restclient.Config) (applications.ServiceRepository, apperrors.AppError) {
	applicationClientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s application client, %s", err)
	}

	rei := applicationClientset.ApplicationconnectorV1alpha1().Applications()

	return applications.NewServiceRepository(rei), nil
}

func newSecretsRepository(coreClientset kubernetes.Interface, namespace string) secrets.Repository {
	sei := coreClientset.CoreV1().Secrets(namespace)

	return secrets.NewRepository(sei)
}

func newCSRFClient(timeout int) csrf.Client {
	cache := csrfClient.NewTokenCache()
	client := &http.Client{}
	return csrfClient.New(timeout, cache, client)
}

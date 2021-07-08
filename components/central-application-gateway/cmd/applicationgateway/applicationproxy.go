package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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
	"github.com/oklog/run"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	shutdownTimeout = 2 * time.Second
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting Application Gateway.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	k8sConfig, err := clientcmd.BuildConfigFromFlags(options.apiServerURL, options.kubeConfig)
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
		Addr:         ":" + strconv.Itoa(options.proxyPort),
		Handler:      internalHandler,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	internalSrvCompass := &http.Server{
		Addr:         ":" + strconv.Itoa(options.proxyPortCompass),
		Handler:      internalHandlerForCompass,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	var g run.Group

	addHttpServerToRunGroup("external-api", &g, externalSrv)
	addHttpServerToRunGroup("proxy-kyma-os", &g, internalSrv)
	addHttpServerToRunGroup("proxy-kyma-mps", &g, internalSrvCompass)
	addInterruptSignalToRunGroup(&g)

	err = g.Run()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func addHttpServerToRunGroup(name string, g *run.Group, srv *http.Server) {
	log.Infof("Starting %s HTTP server on %s", name, srv.Addr)
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatalf("Unable to start %s HTTP server: '%s'", name, err.Error())
	}
	g.Add(func() error {
		defer log.Infof("Server %s finished", name)
		return srv.Serve(ln)
	}, func(error) {
		log.Infof("Shutting down %s HTTP server on %s", name, srv.Addr)

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		err = srv.Shutdown(ctx)
		if err != nil && err != http.ErrServerClosed {
			log.Warnf("HTTP server shutdown %s failed: %s", name, err.Error())
		}
	})
}

func addInterruptSignalToRunGroup(g *run.Group) {
	cancelInterrupt := make(chan struct{})
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-cancelInterrupt:
		case sig := <-c:
			log.Infof("received signal %s", sig)
		}
		return nil
	}, func(error) {
		close(cancelInterrupt)
	})
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
	return csrfClient.New(timeout, cache)
}

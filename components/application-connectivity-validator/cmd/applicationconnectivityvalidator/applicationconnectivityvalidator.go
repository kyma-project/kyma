package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/cache"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/applications"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/externalapi"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/validationproxy"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting Validation Proxy.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	applicationGetter, err := newApplicationGetter()
	if err != nil {
		log.Errorf("Failed to create Application Getter: %s", err)
		os.Exit(1)
	}

	idCache := cache.NewCache(options.cacheExpirationMinutes, options.cacheCleanupMinutes)

	proxyHandler := validationproxy.NewProxyHandler(
		options.group,
		options.tenant,
		options.eventServicePathPrefixV1,
		options.eventServicePathPrefixV2,
		options.eventServiceHost,
		options.appRegistryPathPrefix,
		options.appRegistryHost,
		applicationGetter,
		idCache)

	proxyServer := http.Server{
		Handler: validationproxy.NewHandler(proxyHandler),
		Addr:    fmt.Sprintf(":%d", options.proxyPort),
	}

	externalServer := http.Server{
		Handler: externalapi.NewHandler(),
		Addr:    fmt.Sprintf(":%d", options.externalAPIPort),
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		log.Error(proxyServer.ListenAndServe())
	}()

	go func() {
		log.Error(externalServer.ListenAndServe())
	}()

	wg.Wait()
}

func newApplicationGetter() (applications.Getter, apperrors.AppError) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, apperrors.Internal("failed to get k8s config: %s", err)
	}

	applicationEnvironmentClientset, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s application client: %s", err)
	}

	applicationInterface := applicationEnvironmentClientset.ApplicationconnectorV1alpha1().Applications()
	applicationGetter := applications.NewGetter(applicationInterface)

	return applicationGetter, nil
}

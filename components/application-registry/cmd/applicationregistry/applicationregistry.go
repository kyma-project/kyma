package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/certificates"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/strategy"

	"sync"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/externalapi"
	"github.com/kyma-project/kyma/components/application-registry/internal/httptools"
	"github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/accessservice"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/istio"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/minio"
	metauuid "github.com/kyma-project/kyma/components/application-registry/internal/metadata/uuid"
	"github.com/kyma-project/kyma/components/application-registry/internal/monitoring"
	istioclient "github.com/kyma-project/kyma/components/application-registry/pkg/client/clientset/versioned"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	MinioAccessKeyIdEnv  = "MINIO_ACCESSKEYID"
	MinioSecretAccessKey = "MINIO_ACCESSKEYSECRET"
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting metadata.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	nameResolver := k8sconsts.NewNameResolver(options.namespace)

	serviceDefinitionService, err := newServiceDefinitionService(
		options.minioURL,
		options.namespace,
		options.proxyPort,
		nameResolver,
	)

	if err != nil {
		log.Errorf("Unable to initialize Metadata Service, %s", err.Error())
	}

	middlewares, err := monitoring.SetupMonitoringMiddleware()
	if err != nil {
		log.Errorf("Failed to setup monitoring middleware, %s", err.Error())
	}

	externalHandler := newExternalHandler(serviceDefinitionService, middlewares, options.detailedErrorResponse)

	if options.requestLogging {
		externalHandler = httptools.RequestLogger("External handler: ", externalHandler)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	externalSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(options.externalAPIPort),
		Handler:      externalHandler,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	go func() {
		log.Info(externalSrv.ListenAndServe())
	}()

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Info(http.ListenAndServe(":9090", nil))
	}()

	wg.Wait()
}

func newExternalHandler(serviceDefinitionService metadata.ServiceDefinitionService, middlewares []mux.MiddlewareFunc, detailedErrorResponse bool) http.Handler {
	var metadataHandler externalapi.MetadataHandler

	if serviceDefinitionService != nil {
		metadataHandler = externalapi.NewMetadataHandler(externalapi.NewServiceDetailsValidator(), serviceDefinitionService, detailedErrorResponse)
	} else {
		metadataHandler = externalapi.NewInvalidStateMetadataHandler("Service is not initialized properly.")
	}

	return externalapi.NewHandler(metadataHandler, middlewares)
}

func newServiceDefinitionService(minioURL, namespace string, proxyPort int, nameResolver k8sconsts.NameResolver) (metadata.ServiceDefinitionService, apperrors.AppError) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, apperrors.Internal("Failed to read k8s in-cluster configuration, %s", err)
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("Failed to create k8s core client, %s", err)
	}

	accessKeyId, secretAccessKey, err := readMinioAccessConfiguration()
	if err != nil {
		return nil, apperrors.Internal("Failed to read minio configuration, %s", err.Error())
	}

	minioRepository, err := minio.NewMinioRepository(minioURL, accessKeyId, secretAccessKey)
	if err != nil {
		return nil, apperrors.Internal("Failed to create minio repository, %s", err.Error())
	}

	minioService := minio.NewService(minioRepository)

	specificationService := specification.NewSpecService(minioService)

	applicationServiceRepository, apperror := newApplicationRepository(k8sConfig)
	if apperror != nil {
		return nil, apperror
	}

	istioService, apperror := newIstioService(k8sConfig, namespace)
	if apperror != nil {
		return nil, apperror
	}

	accessServiceManager := newAccessServiceManager(coreClientset, namespace, proxyPort)
	secretsService := newSecretsRepository(coreClientset, nameResolver, namespace)

	uuidGenerator := metauuid.GeneratorFunc(func() string {
		return uuid.NewV4().String()
	})

	serviceAPIService := serviceapi.NewService(nameResolver, accessServiceManager, secretsService, istioService)

	return metadata.NewServiceDefinitionService(uuidGenerator, serviceAPIService, applicationServiceRepository, specificationService), nil
}

func newApplicationRepository(config *restclient.Config) (applications.ServiceRepository, apperrors.AppError) {
	applicationEnvironmentClientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, apperrors.Internal("Failed to create k8s application client, %s", err)
	}

	rei := applicationEnvironmentClientset.ApplicationconnectorV1alpha1().Applications()

	return applications.NewServiceRepository(rei), nil
}

func newAccessServiceManager(coreClientset *kubernetes.Clientset, namespace string, proxyPort int) accessservice.AccessServiceManager {
	si := coreClientset.CoreV1().Services(namespace)

	config := accessservice.AccessServiceManagerConfig{
		TargetPort: int32(proxyPort),
	}

	return accessservice.NewAccessServiceManager(si, config)
}

func newSecretsRepository(coreClientset *kubernetes.Clientset, nameResolver k8sconsts.NameResolver, namespace string) secrets.Service {
	sei := coreClientset.CoreV1().Secrets(namespace)
	strategyFactory := strategy.NewSecretsStrategyFactory(certificates.GenerateKeyAndCertificate)
	repository := secrets.NewRepository(sei)

	return secrets.NewService(repository, nameResolver, strategyFactory)
}

func newIstioService(config *restclient.Config, namespace string) (istio.Service, apperrors.AppError) {
	ic, err := istioclient.NewForConfig(config)
	if err != nil {
		return nil, apperrors.Internal("Failed to create Istio client, %s", err)
	}

	repository := istio.NewRepository(
		ic.IstioV1alpha2().Rules(namespace),
		ic.IstioV1alpha2().Checknothings(namespace),
		ic.IstioV1alpha2().Deniers(namespace),
		istio.RepositoryConfig{Namespace: namespace},
	)

	return istio.NewService(repository), nil
}

func readMinioAccessConfiguration() (string, string, apperrors.AppError) {
	accessKeyId, foundId := os.LookupEnv(MinioAccessKeyIdEnv)
	secretAccessKey, foundSecret := os.LookupEnv(MinioSecretAccessKey)

	if !foundId && !foundSecret {
		return "", "", apperrors.Internal("%s and %s environment variables not found", MinioAccessKeyIdEnv, MinioSecretAccessKey)
	} else if !foundId {
		return "", "", apperrors.Internal("%s environment variable not found", MinioAccessKeyIdEnv)
	} else if !foundSecret {
		return "", "", apperrors.Internal("%s environment variable not found", MinioSecretAccessKey)
	}

	return accessKeyId, secretAccessKey, nil
}

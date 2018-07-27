package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kyma-project/kyma/components/application-connector/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/internal/externalapi"
	"github.com/kyma-project/kyma/components/application-connector/internal/httptools"
	"github.com/kyma-project/kyma/components/application-connector/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/accessservice"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/istio"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/minio"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/secrets"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/serviceapi"
	metauuid "github.com/kyma-project/kyma/components/application-connector/internal/metadata/uuid"
	istioclient "github.com/kyma-project/kyma/components/application-connector/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
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
		log.Errorf("Unable to initialize: '%s'", err.Error())
	}

	externalHandler := newExternalHandler(serviceDefinitionService)

	if options.requestLogging {
		externalHandler = httptools.RequestLogger("External handler: ", externalHandler)
	}

	externalSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(options.externalAPIPort),
		Handler:      externalHandler,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	log.Info(externalSrv.ListenAndServe())
}

func newExternalHandler(serviceDefinitionService metadata.ServiceDefinitionService) http.Handler {
	var metadataHandler externalapi.MetadataHandler

	if serviceDefinitionService != nil {
		metadataHandler = externalapi.NewMetadataHandler(externalapi.NewServiceDetailsValidator(), serviceDefinitionService)
	} else {
		metadataHandler = externalapi.NewInvalidStateMetadataHandler("Service is not initialized properly.")
	}

	return externalapi.NewHandler(metadataHandler)
}

func newServiceDefinitionService(minioURL, namespace string, proxyPort int, nameResolver k8sconsts.NameResolver) (metadata.ServiceDefinitionService, apperrors.AppError) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, apperrors.Internal("failed to read k8s in-cluster configuration, %s", err)
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s core client, %s", err)
	}

	accessKeyId, secretAccessKey, err := readMinioAccessConfiguration()
	if err != nil {
		return nil, apperrors.Internal("failed to read minio configuration, %s", err.Error())
	}

	minioRepository, err := minio.NewMinioRepository(minioURL, accessKeyId, secretAccessKey)
	if err != nil {
		return nil, apperrors.Internal("failed to create minio repository, %s", err.Error())
	}

	minioService := minio.NewService(minioRepository)

	remoteEnvironmentServiceRepository, apperror := newRemoteEnvironmentRepository(k8sConfig)
	if apperror != nil {
		return nil, apperror
	}

	istioService, apperror := newIstioService(k8sConfig, namespace)
	if err != nil {
		return nil, apperror
	}

	accessServiceManager := newAccessServiceManager(coreClientset, namespace, proxyPort)
	secretsRepository := newSecretsRepository(coreClientset, namespace)

	uuidGenerator := metauuid.GeneratorFunc(func() string {
		return uuid.NewV4().String()
	})

	serviceAPIService := serviceapi.NewService(nameResolver, accessServiceManager, secretsRepository, istioService)

	return metadata.NewServiceDefinitionService(uuidGenerator, serviceAPIService, remoteEnvironmentServiceRepository, minioService), nil
}

func newRemoteEnvironmentRepository(config *restclient.Config) (remoteenv.ServiceRepository, apperrors.AppError) {
	remoteEnvironmentClientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s remote environment client, %s", err)
	}

	rei := remoteEnvironmentClientset.RemoteenvironmentV1alpha1().RemoteEnvironments()

	return remoteenv.NewServiceRepository(rei), nil
}

func newAccessServiceManager(coreClientset *kubernetes.Clientset, namespace string, proxyPort int) accessservice.AccessServiceManager {
	si := coreClientset.CoreV1().Services(namespace)

	config := accessservice.AccessServiceManagerConfig{
		TargetPort: int32(proxyPort),
	}

	return accessservice.NewAccessServiceManager(si, config)
}

func newSecretsRepository(coreClientset *kubernetes.Clientset, namespace string) secrets.Repository {
	sei := coreClientset.CoreV1().Secrets(namespace)

	return secrets.NewRepository(sei)
}

func newIstioService(config *restclient.Config, namespace string) (istio.Service, apperrors.AppError) {
	ic, err := istioclient.NewForConfig(config)
	if err != nil {
		return nil, apperrors.Internal("failed to create client for istio, %s", err)
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
		return "", "", apperrors.Internal("%s and %s environment variables not set", MinioAccessKeyIdEnv, MinioSecretAccessKey)
	} else if !foundId {
		return "", "", apperrors.Internal("%s environment variable not set", MinioAccessKeyIdEnv)
	} else if !foundSecret {
		return "", "", apperrors.Internal("%s environment variable not set", MinioSecretAccessKey)
	}

	return accessKeyId, secretAccessKey, nil
}

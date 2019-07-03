package main

import (
	"net/http"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/upload"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/certificates"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/strategy"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/externalapi"
	"github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/accessservice"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/istio"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification"
	metauuid "github.com/kyma-project/kyma/components/application-registry/internal/metadata/uuid"
	istioclient "github.com/kyma-project/kyma/components/application-registry/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

func newExternalHandler(serviceDefinitionService metadata.ServiceDefinitionService, middlewares []mux.MiddlewareFunc, opt *options) http.Handler {
	var metadataHandler externalapi.MetadataHandler

	if serviceDefinitionService != nil {
		metadataHandler = externalapi.NewMetadataHandler(externalapi.NewServiceDetailsValidator(), serviceDefinitionService, opt.detailedErrorResponse)
	} else {
		metadataHandler = externalapi.NewInvalidStateMetadataHandler("Service is not initialized properly.")
	}

	return externalapi.NewHandler(metadataHandler, middlewares)
}

func newServiceDefinitionService(opt *options, nameResolver k8sconsts.NameResolver) (metadata.ServiceDefinitionService, apperrors.AppError) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, apperrors.Internal("Failed to read k8s in-cluster configuration, %s", err)
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("Failed to create k8s core client, %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("Failed to create dynamic client, %s", err)
	}

	specificationService := NewSpecificationService(dynamicClient, opt)

	applicationServiceRepository, apperror := newApplicationRepository(k8sConfig)
	if apperror != nil {
		return nil, apperror
	}

	istioService, apperror := newIstioService(k8sConfig, opt.namespace)
	if apperror != nil {
		return nil, apperror
	}

	sei := coreClientset.CoreV1().Secrets(opt.namespace)
	secretsRepository := secrets.NewRepository(sei)

	accessServiceManager := newAccessServiceManager(coreClientset, opt.namespace, opt.proxyPort)
	credentialsSecretsService := newSecretsService(secretsRepository, nameResolver)
	requestParametersSecretsService := secrets.NewRequestParametersService(secretsRepository, nameResolver)

	uuidGenerator := metauuid.GeneratorFunc(func() (string, error) {
		uuidInstance, err := uuid.NewV4()
		if err != nil {
			return "", err
		}
		return uuidInstance.String(), nil
	})

	serviceAPIService := serviceapi.NewService(nameResolver, accessServiceManager, credentialsSecretsService, requestParametersSecretsService, istioService)

	return metadata.NewServiceDefinitionService(uuidGenerator, serviceAPIService, applicationServiceRepository, specificationService), nil
}

func NewSpecificationService(dynamicClient dynamic.Interface, opt *options) specification.Service {
	groupVersionResource := schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "clusterdocstopics",
	}
	resourceInterface := dynamicClient.Resource(groupVersionResource).Namespace(opt.namespace)

	docsTopicRepository := assetstore.NewDocsTopicRepository(resourceInterface)
	uploadClient := upload.NewClient(opt.uploadServiceURL)
	assetStoreService := assetstore.NewService(docsTopicRepository, uploadClient, opt.insecureAssetDownload, opt.assetstoreRequestTimeout)

	return specification.NewSpecService(assetStoreService, opt.specRequestTimeout, opt.insecureSpecDownload)
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

func newSecretsService(repository secrets.Repository, nameResolver k8sconsts.NameResolver) secrets.Service {
	strategyFactory := strategy.NewSecretsStrategyFactory(certificates.GenerateKeyAndCertificate)

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

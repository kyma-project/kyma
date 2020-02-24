package main

import (
	appclient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	istioclient "github.com/kyma-project/kyma/components/application-registry/pkg/client/clientset/versioned"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/k8sconsts"
	"kyma-project.io/compass-runtime-agent/internal/kyma"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/accessservice"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/istio"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/upload"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/secrets"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/secrets/strategy"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications"
	"kyma-project.io/compass-runtime-agent/internal/metrics"
	"time"
)

type k8sResourceClientSets struct {
	core        *kubernetes.Clientset
	istio       *istioclient.Clientset
	application *appclient.Clientset
	dynamic     dynamic.Interface
}

func k8sResourceClients(k8sConfig *restclient.Config) (*k8sResourceClientSets, error) {
	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create k8s core client")
	}

	istioClientset, err := istioclient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("Failed to create Istio client, %s", err)
	}

	applicationClientset, err := appclient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("Failed to create k8s application client, %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("Failed to create dynamic client, %s", err)
	}

	return &k8sResourceClientSets{
		core:        coreClientset,
		istio:       istioClientset,
		application: applicationClientset,
		dynamic:     dynamicClient,
	}, nil
}

func createNewSynchronizationService(k8sResourceClients *k8sResourceClientSets, secretsManager secrets.Manager, namespace string, gatewayPort int, uploadServiceUrl string) (kyma.Service, error) {
	nameResolver := k8sconsts.NewNameResolver(namespace)
	converter := applications.NewConverter(nameResolver)

	applicationManager := newApplicationManager(k8sResourceClients.application)
	accessServiceManager := newAccessServiceManager(k8sResourceClients.core, namespace, gatewayPort)
	istioService := newIstioService(k8sResourceClients.istio, namespace)

	resourcesService := newResourcesService(secretsManager, accessServiceManager, istioService, k8sResourceClients.dynamic, nameResolver, uploadServiceUrl)

	return kyma.NewService(applicationManager, converter, resourcesService), nil
}

func newResourcesService(secretsManager secrets.Manager, accessServiceMgr accessservice.AccessServiceManager, istioSvc istio.Service,
	dynamicClient dynamic.Interface, nameResolver k8sconsts.NameResolver, uploadServiceUrl string) apiresources.Service {

	secretsRepository := secrets.NewRepository(secretsManager)

	secretsService := newSecretsService(secretsRepository, nameResolver)

	rafterService := newRafter(dynamicClient, uploadServiceUrl)

	return apiresources.NewService(accessServiceMgr, secretsService, nameResolver, istioSvc, rafterService)
}

func newRafter(dynamicClient dynamic.Interface, uploadServiceURL string) rafter.Service {
	groupVersionResource := v1beta1.GroupVersion.WithResource("clusterassetgroups")
	resourceInterface := dynamicClient.Resource(groupVersionResource)

	clusterAssetGroupRepository := rafter.NewAssetGroupRepository(resourceInterface)
	uploadClient := upload.NewClient(uploadServiceURL)
	return rafter.NewService(clusterAssetGroupRepository, uploadClient)
}

func newAccessServiceManager(coreClientset *kubernetes.Clientset, namespace string, proxyPort int) accessservice.AccessServiceManager {
	si := coreClientset.CoreV1().Services(namespace)

	config := accessservice.AccessServiceManagerConfig{
		TargetPort: int32(proxyPort),
	}

	return accessservice.NewAccessServiceManager(si, config)
}

func newApplicationManager(appClientset *appclient.Clientset) applications.Repository {
	appInterface := appClientset.ApplicationconnectorV1alpha1().Applications()
	return applications.NewRepository(appInterface)
}

func newSecretsService(repository secrets.Repository, nameResolver k8sconsts.NameResolver) secrets.Service {
	strategyFactory := strategy.NewSecretsStrategyFactory()

	return secrets.NewService(repository, nameResolver, strategyFactory)
}

func newIstioService(ic *istioclient.Clientset, namespace string) istio.Service {
	repository := istio.NewRepository(
		ic.IstioV1alpha2().Rules(namespace),
		ic.IstioV1alpha2().Instances(namespace),
		ic.IstioV1alpha2().Handlers(namespace),
		istio.RepositoryConfig{Namespace: namespace},
	)

	return istio.NewService(repository)
}

func newMetricsLogger(loggingTimeInterval time.Duration) (metrics.Logger, error) {
	config, err := restclient.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster config")
	}

	resourcesClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create resources clientset for config")
	}

	metricsClientset, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create metrics clientset for config")
	}

	return metrics.NewMetricsLogger(resourcesClientset, metricsClientset, loggingTimeInterval), nil
}

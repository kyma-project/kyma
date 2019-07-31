package main

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/accessservice"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/strategy"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

func createNewSynchronizationService() kyma.Service {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.Errorf("Failed to read k8s in-cluster configuration, %s", err)

		return uninitializedKymaService{}
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Errorf("Failed to create k8s core client, %s", err)
		return uninitializedKymaService{}
	}

	applicationManager, err := newApplicationManager(k8sConfig)
	if err != nil {
		log.Errorf("Failed to initialize Applications manager, %s", err)
		return uninitializedKymaService{}
	}

	// TODO: pass the namespace name in parameters
	nameResolver := k8sconsts.NewNameResolver("kyma-integration")
	converter := applications.NewConverter(nameResolver)

	resourcesService := newResourcesService(coreClientset, nameResolver)

	return kyma.NewService(applicationManager, converter, resourcesService)
}

func newResourcesService(coreClientset *kubernetes.Clientset, nameResolver k8sconsts.NameResolver) apiresources.Service {
	// TODO: pass proxy port and the namespace in parameters
	accessServiceManager := newAccessServiceManager(coreClientset, "kyma-integration", 8080)

	sei := coreClientset.CoreV1().Secrets("kyma-integration")
	secretsRepository := secrets.NewRepository(sei)

	secretsService := newSecretsService(secretsRepository, nameResolver)
	return apiresources.NewService(accessServiceManager, secretsService, nameResolver)
}

func newAccessServiceManager(coreClientset *kubernetes.Clientset, namespace string, proxyPort int) accessservice.AccessServiceManager {
	si := coreClientset.CoreV1().Services(namespace)

	config := accessservice.AccessServiceManagerConfig{
		TargetPort: int32(proxyPort),
	}

	return accessservice.NewAccessServiceManager(si, config)
}

func newApplicationManager(config *restclient.Config) (applications.Repository, apperrors.AppError) {
	applicationEnvironmentClientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, apperrors.Internal("Failed to create k8s application client, %s", err)
	}

	appInterface := applicationEnvironmentClientset.ApplicationconnectorV1alpha1().Applications()

	return applications.NewRepository(appInterface), nil
}

func newSecretsService(repository secrets.Repository, nameResolver k8sconsts.NameResolver) secrets.Service {
	strategyFactory := strategy.NewSecretsStrategyFactory()

	return secrets.NewService(repository, nameResolver, strategyFactory)
}

type uninitializedKymaService struct {
}

func (u uninitializedKymaService) Apply(applications []model.Application) ([]kyma.Result, apperrors.AppError) {
	return nil, apperrors.Internal("Service not initialized")
}

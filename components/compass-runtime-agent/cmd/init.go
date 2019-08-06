package main

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	istioclient "github.com/kyma-project/kyma/components/application-registry/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/accessservice"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/istio"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/strategy"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

func createNewSynchronizationService(k8sConfig *restclient.Config, namespace string, gatewayPort int) (kyma.Service, error) {
	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create k8s core client")
	}

	ic, err := istioclient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("Failed to create Istio client, %s", err)
	}

	applicationManager, err := newApplicationManager(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize Applications manager")
	}

	nameResolver := k8sconsts.NewNameResolver(namespace)
	converter := applications.NewConverter(nameResolver)

	resourcesService := newResourcesService(coreClientset, ic, nameResolver, namespace, gatewayPort)

	return kyma.NewService(applicationManager, converter, resourcesService), nil
}

func newResourcesService(coreClientset *kubernetes.Clientset, ic *istioclient.Clientset, nameResolver k8sconsts.NameResolver, namespace string, gatewayPort int) apiresources.Service {
	accessServiceManager := newAccessServiceManager(coreClientset, namespace, gatewayPort)

	sei := coreClientset.CoreV1().Secrets(namespace)
	secretsRepository := secrets.NewRepository(sei)

	secretsService := newSecretsService(secretsRepository, nameResolver)

	istioService := newIstioService(ic, namespace)

	return apiresources.NewService(accessServiceManager, secretsService, nameResolver, istioService)
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

func newIstioService(ic *istioclient.Clientset, namespace string) istio.Service {
	repository := istio.NewRepository(
		ic.IstioV1alpha2().Rules(namespace),
		ic.IstioV1alpha2().Instances(namespace),
		ic.IstioV1alpha2().Handlers(namespace),
		istio.RepositoryConfig{Namespace: namespace},
	)

	return istio.NewService(repository)
}

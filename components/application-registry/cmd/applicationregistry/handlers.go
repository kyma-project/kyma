package main

import (
	v1alpha12 "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/certificates"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/strategy"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/accessservice"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

func newApplicationManager(config *restclient.Config) (v1alpha12.ApplicationInterface, apperrors.AppError) {
	applicationEnvironmentClientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, apperrors.Internal("Failed to create k8s application client, %s", err)
	}

	return applicationEnvironmentClientset.ApplicationconnectorV1alpha1().Applications(), nil
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
add
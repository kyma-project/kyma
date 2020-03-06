package resource_watcher

import (
	"time"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	ConfigLabel                   = "serverless.kyma-project.io/config"
	RuntimeLabel                  = "serverless.kyma-project.io/runtime"
	RegistryCredentialsLabelValue = "registry-credentials"
	ServiceAccountLabelValue      = "service-account"
	RuntimeLabelValue             = "runtime"
)

type Config struct {
	EnableControllers       bool          `default:"true"`
	BaseNamespace           string        `default:"kyma-system"`
	ExcludedNamespaces      []string      `default:"kube-system,kyma-system"`
	NamespaceRelistInterval time.Duration `default:"60s"`
}

type Services struct {
	Namespaces     *NamespaceService
	Credentials    *CredentialsService
	Runtimes       *RuntimesService
	ServiceAccount *ServiceAccountService
}

func NewResourceWatcherServices(coreClient *v1.CoreV1Client, config Config) *Services {
	namespacesServices := NewNamespaceService(coreClient, config)
	credentialsServices := NewCredentialsService(coreClient, config)
	runtimesServices := NewRuntimesService(coreClient, config)
	serviceAccountsServices := NewServiceAccountService(coreClient, config, credentialsServices)

	return &Services{
		Namespaces:     namespacesServices,
		Credentials:    credentialsServices,
		Runtimes:       runtimesServices,
		ServiceAccount: serviceAccountsServices,
	}
}

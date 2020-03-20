package configwatcher

import (
	"time"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	ConfigLabel              = "serverless.kyma-project.io/config"
	CredentialsLabel         = "serverless.kyma-project.io/credentials"
	RuntimeLabel             = "serverless.kyma-project.io/runtime"
	CredentialsLabelValue    = "credentials"
	ServiceAccountLabelValue = "service-account"
	RuntimeLabelValue        = "runtime"
)

type Config struct {
	EnableControllers       bool          `default:"true"`
	BaseNamespace           string        `default:"kyma-system"`
	ExcludedNamespaces      []string      `default:"istio-system,knative-eventing,knative-serving,kube-node-lease,kube-public,kube-system,kyma-installer,kyma-integration,kyma-system,tekton-pipelines,natss"`
	NamespaceRelistInterval time.Duration `default:"1m"`
}

type Services struct {
	Namespaces     *NamespaceService
	Credentials    *CredentialsService
	Runtimes       *RuntimesService
	ServiceAccount *ServiceAccountService
}

func NewConfigWatcherServices(coreClient v1.CoreV1Interface, config Config) *Services {
	namespacesServices := NewNamespaceService(coreClient, config)
	credentialsServices := NewCredentialsService(coreClient, config)
	runtimesServices := NewRuntimesService(coreClient, config)
	serviceAccountsServices := NewServiceAccountService(coreClient, config)

	return &Services{
		Namespaces:     namespacesServices,
		Credentials:    credentialsServices,
		Runtimes:       runtimesServices,
		ServiceAccount: serviceAccountsServices,
	}
}

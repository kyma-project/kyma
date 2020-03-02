package resource_watcher

import (
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	ConfigLabel = "serverless.kyma-project.io/config"
	RuntimeLabel = "serverless.kyma-project.io/runtime"
)

type ResourceWatcherConfig struct {
	EnableControllers bool `envconfig:"default=true"`
	BaseNamespace         string   `envconfig:"default=kyma-system"`
	ExcludedNamespaces    []string `envconfig:"default=kube-system,kyma-system"`
}

type ResourceWatcherServices struct {
	Namespaces *NamespaceService
	Credentials *CredentialsService
	Runtimes *RuntimesService
}

func NewResourceWatcherServices(coreClient *v1.CoreV1Client, config ResourceWatcherConfig) *ResourceWatcherServices {
	return &ResourceWatcherServices{
		Namespaces: NewNamespaceService(coreClient, config),
		Credentials: NewCredentialsService(coreClient, config.BaseNamespace),
		Runtimes: NewRuntimesService(coreClient, config.BaseNamespace),
	}
}
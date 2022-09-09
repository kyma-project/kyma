package compassruntimeagentinit

import "k8s.io/client-go/kubernetes"

type RollbackDeploymentFunc func() error

type DeploymentConfigurator interface {
	Do(name, namespace string) (RollbackDeploymentFunc, error)
}

type deploymentConfiguration struct {
	coreClientset *kubernetes.Clientset
}

func newDeploymentConfiguration(coreClientset *kubernetes.Clientset) DeploymentConfigurator {
	return deploymentConfiguration{
		coreClientset: coreClientset,
	}
}

func (d deploymentConfiguration) Do(name, namespace string) (RollbackDeploymentFunc, error) {
	return nil, nil
}

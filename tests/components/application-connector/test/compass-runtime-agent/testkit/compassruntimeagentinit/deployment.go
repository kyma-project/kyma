package compassruntimeagentinit

import "k8s.io/client-go/kubernetes"

type RollbackDeploymentFunc func() error

type DeploymentConfigurator interface {
	Do(name, namespace string) (RollbackDeploymentFunc, error)
}

type deploymentConfiguration struct {
	kubernetesInterface kubernetes.Interface
}

func newDeploymentConfiguration(kubernetesInterface kubernetes.Interface) DeploymentConfigurator {
	return deploymentConfiguration{
		kubernetesInterface: kubernetesInterface,
	}
}

func (d deploymentConfiguration) Do(name, namespace string) (RollbackDeploymentFunc, error) {
	return nil, nil
}

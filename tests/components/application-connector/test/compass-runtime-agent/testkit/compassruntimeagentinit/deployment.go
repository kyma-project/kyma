package compassruntimeagentinit

import "k8s.io/client-go/kubernetes"

type RollbackDeploymentFunc func() error

type deploymentConfiguration struct {
	kubernetesInterface kubernetes.Interface
}

func newDeploymentConfiguration(kubernetesInterface kubernetes.Interface) deploymentConfiguration {
	return deploymentConfiguration{
		kubernetesInterface: kubernetesInterface,
	}
}

func (d deploymentConfiguration) Do(name, namespace string) (RollbackDeploymentFunc, error) {
	return nil, nil
}

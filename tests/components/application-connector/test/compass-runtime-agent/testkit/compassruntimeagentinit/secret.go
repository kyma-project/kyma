package compassruntimeagentinit

import "k8s.io/client-go/kubernetes"

type RollbackSecretFunc func() error

type SecretCreator interface {
	Do(name, namespace string, compassRuntimeAgentConfig CompassRuntimeAgentConfig) (RollbackSecretFunc, error)
}

type secretCreator struct {
	kubernetesInterface kubernetes.Interface
}

func newSecretCreator(kubernetesInterface kubernetes.Interface) SecretCreator {
	return secretCreator{
		kubernetesInterface: kubernetesInterface,
	}
}

func (s secretCreator) Do(name, namespace string, compassRuntimeAgentConfig CompassRuntimeAgentConfig) (RollbackSecretFunc, error) {
	return nil, nil
}

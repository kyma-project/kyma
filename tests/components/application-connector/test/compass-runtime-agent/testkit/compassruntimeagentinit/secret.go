package compassruntimeagentinit

import "k8s.io/client-go/kubernetes"

type RollbackSecretFunc func() error

type SecretCreator interface {
	Do(name, namespace string, compassRuntimeAgentConfig CompassRuntimeAgentConfig) (RollbackSecretFunc, error)
}

type secretCreator struct {
	coreClientset *kubernetes.Clientset
}

func newSecretCreator(coreClientset *kubernetes.Clientset) SecretCreator {
	return secretCreator{
		coreClientset: coreClientset,
	}
}

func (s secretCreator) Do(name, namespace string, compassRuntimeAgentConfig CompassRuntimeAgentConfig) (RollbackSecretFunc, error) {
	return nil, nil
}

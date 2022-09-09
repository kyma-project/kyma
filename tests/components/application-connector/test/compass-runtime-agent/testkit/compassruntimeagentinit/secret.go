package compassruntimeagentinit

import (
	"context"
	"github.com/avast/retry-go"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

type RollbackSecretFunc func() error

const (
	connectorURLConfigKey = "CONNECTOR_URL"
	tokenConfigKey        = "TOKEN"
	runtimeIdConfigKey    = "RUNTIME_ID"
	tenantConfigKey       = "TENANT"
)

type secretCreator struct {
	kubernetesInterface kubernetes.Interface
}

func newSecretCreator(kubernetesInterface kubernetes.Interface) secretCreator {
	return secretCreator{
		kubernetesInterface: kubernetesInterface,
	}
}

func (s secretCreator) Do(name, namespace string, config CompassRuntimeAgentConfig) (RollbackSecretFunc, error) {

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			connectorURLConfigKey: []byte(config.ConnectorUrl),
			tokenConfigKey:        []byte(config.Token),
			runtimeIdConfigKey:    []byte(config.RuntimeID),
			tenantConfigKey:       []byte(config.Tenant),
		},
	}

	// TODO: Adjust retry functionality (attempts count, etc.). Maybe it would be good to make it configurable
	// TODO: consider if we need to add retry at this point
	err := retry.Do(func() error {
		_, err := s.kubernetesInterface.CoreV1().Secrets(namespace).Create(context.Background(), &secret, metav1.CreateOptions{})
		if err != nil {
			return err
		}

		return nil
	}, retry.Attempts(3), retry.MaxDelay(5*time.Second))

	if err != nil {
		return nil, err
	}

	return s.newRollbackSecretFunc(name, namespace), nil
}

func (s secretCreator) newRollbackSecretFunc(name, namespace string) RollbackSecretFunc {
	return func() error {
		return retry.Do(func() error {
			return s.kubernetesInterface.CoreV1().Secrets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
		}, retry.Attempts(3), retry.MaxDelay(5*time.Second))
	}
}

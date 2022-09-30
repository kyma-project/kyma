package compassruntimeagentinit

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit/types"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	connectorURLConfigKey = "CONNECTOR_URL"
	tokenConfigKey        = "TOKEN"
	runtimeIdConfigKey    = "RUNTIME_ID"
	tenantConfigKey       = "TENANT"
)

type configurationSecretConfigurator struct {
	kubernetesInterface          kubernetes.Interface
	configurationSecretNamespace string
}

func NewConfigurationSecretConfigurator(kubernetesInterface kubernetes.Interface, configurationSecretNamespace string) configurationSecretConfigurator {
	return configurationSecretConfigurator{
		kubernetesInterface:          kubernetesInterface,
		configurationSecretNamespace: configurationSecretNamespace,
	}
}

func (s configurationSecretConfigurator) Do(newConfigSecretName string, config types.CompassRuntimeAgentConfig) (types.RollbackFunc, error) {

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newConfigSecretName,
			Namespace: s.configurationSecretNamespace,
		},
		Data: map[string][]byte{
			connectorURLConfigKey: []byte(config.ConnectorUrl),
			tokenConfigKey:        []byte(config.Token),
			runtimeIdConfigKey:    []byte(config.RuntimeID),
			tenantConfigKey:       []byte(config.Tenant),
		},
	}

	err := retry.Do(func() error {
		_, err := s.kubernetesInterface.CoreV1().Secrets(s.configurationSecretNamespace).Create(context.Background(), &secret, metav1.CreateOptions{})
		if err != nil {
			return err
		}

		return nil
	}, retry.Attempts(RetryAttempts), retry.Delay(RetrySeconds*time.Second))

	if err != nil {
		return nil, err
	}

	return s.newRollbackSecretFunc(newConfigSecretName, s.configurationSecretNamespace), nil
}

func (s configurationSecretConfigurator) newRollbackSecretFunc(name, namespace string) types.RollbackFunc {
	return func() error {
		return deleteSecretWithRetry(s.kubernetesInterface, name, namespace)
	}
}

func deleteSecretWithRetry(kubernetesInterface kubernetes.Interface, name, namespace string) error {
	return retry.Do(func() error {
		err := kubernetesInterface.CoreV1().Secrets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}
		}

		return err
	}, retry.Attempts(RetryAttempts), retry.Delay(RetrySeconds*time.Second))
}

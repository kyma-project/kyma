package compassruntimeagentinit

import (
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit/types"
	"k8s.io/client-go/kubernetes"
)

type certificatesSecretsConfigurator struct {
	kubernetesInterface kubernetes.Interface
	namespace           string
}

func NewCertificateSecretConfigurator(kubernetesInterface kubernetes.Interface, namespace string) certificatesSecretsConfigurator {
	return certificatesSecretsConfigurator{
		kubernetesInterface: kubernetesInterface,
		namespace:           namespace,
	}
}

func (csc certificatesSecretsConfigurator) Do(caSecretName, clusterCertSecretName string) (types.RollbackFunc, error) {
	// Original secrets created by Compass Runtime Agent are left intact so that they can be restored after the test.
	// As part of the test preparation new secret names are passed to the Compass Runtime Agent Deployment. Rollback function needs to delete those.
	return csc.getRollbackFunction(caSecretName, clusterCertSecretName), nil
}

func (csc certificatesSecretsConfigurator) getRollbackFunction(caSecretName, clusterCertSecretName string) types.RollbackFunc {
	return func() error {
		var result *multierror.Error

		err := deleteSecretWithRetry(csc.kubernetesInterface, caSecretName, csc.namespace)
		if err != nil {
			multierror.Append(result, err)
		}

		err = deleteSecretWithRetry(csc.kubernetesInterface, clusterCertSecretName, csc.namespace)
		if err != nil {
			multierror.Append(result, err)
		}

		return result.ErrorOrNil()
	}
}

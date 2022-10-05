package init

import (
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init/types"
	"k8s.io/client-go/kubernetes"
)

type certificatesSecretsConfigurator struct {
	kubernetesInterface kubernetes.Interface
}

func NewCertificateSecretConfigurator(kubernetesInterface kubernetes.Interface) certificatesSecretsConfigurator {
	return certificatesSecretsConfigurator{
		kubernetesInterface: kubernetesInterface,
	}
}

func (csc certificatesSecretsConfigurator) Do(newCASecretName, newClusterCertSecretName string) (types.RollbackFunc, error) {
	// Original secrets created by Compass Runtime Agent are left intact so that they can be restored after the test.
	// As part of the test preparation new secret names are passed to the Compass Runtime Agent Deployment. Rollback function needs to delete those.
	return csc.getRollbackFunction(newCASecretName, newClusterCertSecretName), nil
}

func (csc certificatesSecretsConfigurator) getRollbackFunction(caSecretName, clusterCertSecretName string) types.RollbackFunc {
	return func() error {
		var result *multierror.Error

		err := deleteSecretWithRetry(csc.kubernetesInterface, caSecretName, IstioSystemNamespace)
		if err != nil {
			multierror.Append(result, err)
		}

		err = deleteSecretWithRetry(csc.kubernetesInterface, clusterCertSecretName, CompassSystemNamespace)
		if err != nil {
			multierror.Append(result, err)
		}

		return result.ErrorOrNil()
	}
}

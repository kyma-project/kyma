package init

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init/types"
	"github.com/pkg/errors"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v13 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"time"
)

const (
	CRAContainerNumber          = 0
	ConfigurationSecretEnvName  = "APP_AGENT_CONFIGURATION_SECRET"
	CASecretEnvName             = "APP_CA_CERTIFICATES_SECRET"
	ClusterCertSecretEnvName    = "APP_CLUSTER_CERTIFICATES_SECRET"
	ControllerSyncPeriodEnvTime = "APP_CONTROLLER_SYNC_PERIOD"
)

type deploymentConfiguration struct {
	kubernetesInterface kubernetes.Interface
	deploymentName      string
	namespaceName       string
}

func NewDeploymentConfiguration(kubernetesInterface kubernetes.Interface, deploymentName, namespaceName string) deploymentConfiguration {
	return deploymentConfiguration{
		kubernetesInterface: kubernetesInterface,
		deploymentName:      deploymentName,
		namespaceName:       namespaceName,
	}
}

func (dc deploymentConfiguration) Do(newCANamespacedSecretName, newClusterNamespacedCertSecretName, newConfigNamespacedSecretName, newControllerSyncPeriodTime string) (types.RollbackFunc, error) {
	deploymentInterface := dc.kubernetesInterface.AppsV1().Deployments(dc.namespaceName)

	deployment, err := retryGetDeployment(dc.deploymentName, deploymentInterface)
	if err != nil {
		return nil, err
	}

	if len(deployment.Spec.Template.Spec.Containers) < 1 {
		return nil, fmt.Errorf("no containers found in %s/%s deployment", "kyma-system", dc.deploymentName)
	}

	previousConfigSecretNamespacedName, found := replaceEnvValue(deployment, ConfigurationSecretEnvName, newConfigNamespacedSecretName)
	if !found {
		return nil, fmt.Errorf("environment variable '%s' not found in %s deployment", ConfigurationSecretEnvName, dc.deploymentName)
	}

	previousCASecretNamespacedName, found := replaceEnvValue(deployment, CASecretEnvName, newCANamespacedSecretName)
	if !found {
		return nil, fmt.Errorf("environment variable '%s' not found in %s deployment", CASecretEnvName, dc.deploymentName)
	}

	previousCertSecretNamespacedName, found := replaceEnvValue(deployment, ClusterCertSecretEnvName, newClusterNamespacedCertSecretName)
	if !found {
		return nil, fmt.Errorf("environment variable '%s' not found in %s deployment", ClusterCertSecretEnvName, dc.deploymentName)
	}

	previousControllerSyncPeriodTime, found := replaceEnvValue(deployment, ControllerSyncPeriodEnvTime, newControllerSyncPeriodTime)
	if !found {
		return nil, fmt.Errorf("environment variable '%s' not found in %s deployment", ControllerSyncPeriodEnvTime, dc.deploymentName)
	}

	err = retryUpdateDeployment(deployment, deploymentInterface)
	if err != nil {
		return nil, err
	}
	rollbackDeploymentFunc := newRollbackDeploymentFunc(dc.deploymentName, previousConfigSecretNamespacedName, previousCASecretNamespacedName, previousCertSecretNamespacedName, previousControllerSyncPeriodTime, deploymentInterface)

	err = waitForRollout(dc.deploymentName, deploymentInterface)

	return rollbackDeploymentFunc, err
}

func newRollbackDeploymentFunc(name, previousConfigSecretNamespacedName, previousCASecretNamespacedName, previousCertSecretNamespacedName, previousControllerSyncPeriodTime string, deploymentInterface v13.DeploymentInterface) types.RollbackFunc {
	return func() error {
		deployment, err := retryGetDeployment(name, deploymentInterface)
		if err != nil {
			return err
		}

		_, found := replaceEnvValue(deployment, ConfigurationSecretEnvName, previousConfigSecretNamespacedName)
		if !found {
			return fmt.Errorf("environment variable '%s' not found in %s deployment", ConfigurationSecretEnvName, name)
		}

		_, found = replaceEnvValue(deployment, CASecretEnvName, previousCASecretNamespacedName)
		if !found {
			return fmt.Errorf("environment variable '%s' not found in %s deployment", CASecretEnvName, name)
		}

		_, found = replaceEnvValue(deployment, ClusterCertSecretEnvName, previousCertSecretNamespacedName)
		if !found {
			return fmt.Errorf("environment variable '%s' not found in %s deployment", ClusterCertSecretEnvName, name)
		}

		_, found = replaceEnvValue(deployment, ControllerSyncPeriodEnvTime, previousControllerSyncPeriodTime)
		if !found {
			return fmt.Errorf("environment variable '%s' not found in %s deployment", ControllerSyncPeriodEnvTime, name)
		}

		return retryUpdateDeployment(deployment, deploymentInterface)
	}
}

func replaceEnvValue(deployment *v12.Deployment, name, newValue string) (string, bool) {
	envs := deployment.Spec.Template.Spec.Containers[CRAContainerNumber].Env
	for i := range envs {
		if envs[i].Name == name {
			previousValue := envs[i].Value
			envs[i].Value = newValue
			deployment.Spec.Template.Spec.Containers[CRAContainerNumber].Env = envs

			return previousValue, true
		}
	}

	return "", false
}

func retryGetDeployment(name string, deploymentInterface v13.DeploymentInterface) (*v12.Deployment, error) {
	var deployment *v12.Deployment

	err := retry.Do(func() error {
		var err error
		deployment, err = deploymentInterface.Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to get Compass Runtime Agent deployment")
		}
		return nil
	}, retry.Attempts(RetryAttempts), retry.Delay(RetrySeconds*time.Second))

	return deployment, err
}

func retryUpdateDeployment(deployment *v12.Deployment, deploymentInterface v13.DeploymentInterface) error {
	return retry.Do(func() error {
		_, err := deploymentInterface.Update(context.TODO(), deployment, v1.UpdateOptions{})
		return errors.Wrap(err, "failed to update Compass Runtime Agent deployment")
	}, retry.Attempts(RetryAttempts), retry.Delay(RetrySeconds*time.Second))
}

func waitForRollout(name string, deploymentInterface v13.DeploymentInterface) error {
	return retry.Do(func() error {
		deployment, err := deploymentInterface.Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to get Compass Runtime Agent deployment")
		}
		if deployment.Status.AvailableReplicas == 0 || deployment.Status.UnavailableReplicas != 0 {
			return fmt.Errorf("deployment %s is not yet ready", name)
		}
		return nil
	}, retry.Attempts(RetryAttempts), retry.Delay(RetrySeconds*time.Second))
}

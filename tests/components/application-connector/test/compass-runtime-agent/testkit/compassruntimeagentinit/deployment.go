package compassruntimeagentinit

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	types "github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit/types"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v13 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"time"
)

const (
	CRAContainerNumber         = 0
	ConfigurationSecretEnvName = "APP_AGENT_CONFIGURATION_SECRET"
)

type deploymentConfiguration struct {
	kubernetesInterface kubernetes.Interface
}

func NewDeploymentConfiguration(kubernetesInterface kubernetes.Interface) deploymentConfiguration {
	return deploymentConfiguration{
		kubernetesInterface: kubernetesInterface,
	}
}

func (dc deploymentConfiguration) Do(deploymentName, secretName, namespace string) (types.RollbackFunc, error) {
	deploymentInterface := dc.kubernetesInterface.AppsV1().Deployments(namespace)

	deployment, err := retryGetDeployment(deploymentName, deploymentInterface)
	if err != nil {
		return nil, err
	}

	if len(deployment.Spec.Template.Spec.Containers) < 1 {
		return nil, fmt.Errorf("no containers found in %s/%s deployment", namespace, deploymentName)
	}
	envs := deployment.Spec.Template.Spec.Containers[CRAContainerNumber].Env
	previousSecretNamespacedName := ""
	for i := range envs {
		if envs[i].Name == ConfigurationSecretEnvName {
			previousSecretNamespacedName = envs[i].Value
			envs[i].Value = fmt.Sprintf("%s/%s", namespace, secretName)
			break
		}
	}
	if previousSecretNamespacedName == "" {
		return nil, fmt.Errorf("no %s environment variable found in %s/%s deployment", ConfigurationSecretEnvName, namespace, deploymentName)
	}
	deployment.Spec.Template.Spec.Containers[CRAContainerNumber].Env = envs

	if err := retry.Do(func() error {
		_, err := deploymentInterface.Update(context.TODO(), deployment, v1.UpdateOptions{})
		return err
	}, retry.Attempts(RetryAttempts), retry.Delay(RetrySeconds*time.Second)); err != nil {
		return nil, err
	}

	err = waitForRollout(deploymentName, deploymentInterface)
	rollbackDeploymentFunc := newRollbackDeploymentFunc(deploymentName, previousSecretNamespacedName, deploymentInterface)

	return rollbackDeploymentFunc, err
}

func retryGetDeployment(name string, deploymentInterface v13.DeploymentInterface) (*v12.Deployment, error) {
	var deployment *v12.Deployment
	return deployment, retry.Do(func() error {
		var err error
		deployment, err = deploymentInterface.Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(RetryAttempts), retry.Delay(RetrySeconds*time.Second))
}

func waitForRollout(name string, deploymentInterface v13.DeploymentInterface) error {
	return retry.Do(func() error {
		deployment, err := deploymentInterface.Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return err
		}
		if deployment.Status.AvailableReplicas == 0 || deployment.Status.UnavailableReplicas != 0 {
			return fmt.Errorf("deployment %s is not yet ready", name)
		}
		return nil
	}, retry.Attempts(RetryAttempts), retry.Delay(RetrySeconds*time.Second))
}

func newRollbackDeploymentFunc(name, previousSecretNamespacedName string, deploymentInterface v13.DeploymentInterface) types.RollbackFunc {
	return func() error {
		deployment, err := retryGetDeployment(name, deploymentInterface)
		if err != nil {
			return err
		}

		if len(deployment.Spec.Template.Spec.Containers) < 1 {
			return fmt.Errorf("no containers found in %s deployment", name)
		}
		envs := deployment.Spec.Template.Spec.Containers[CRAContainerNumber].Env
		foundEnv := false
		for i := range envs {
			if envs[i].Name == ConfigurationSecretEnvName {
				foundEnv = true
				envs[i].Value = previousSecretNamespacedName
				break
			}
		}
		if foundEnv == false {
			return fmt.Errorf("no %s environment variable found in %s deployment", ConfigurationSecretEnvName, name)
		}
		deployment.Spec.Template.Spec.Containers[CRAContainerNumber].Env = envs

		return retry.Do(func() error {
			_, err := deploymentInterface.Update(context.TODO(), deployment, v1.UpdateOptions{})
			return err
		}, retry.Attempts(RetryAttempts), retry.Delay(RetrySeconds*time.Second))
	}
}

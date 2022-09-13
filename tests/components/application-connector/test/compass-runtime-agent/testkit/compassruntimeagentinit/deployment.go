package compassruntimeagentinit

import (
	"context"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const ConfigurationSecretEnvName = "APP_AGENT_CONFIGURATION_SECRET"

type RollbackDeploymentFunc func() error

type deploymentConfiguration struct {
	kubernetesInterface kubernetes.Interface
}

func newDeploymentConfiguration(kubernetesInterface kubernetes.Interface) deploymentConfiguration {
	return deploymentConfiguration{
		kubernetesInterface: kubernetesInterface,
	}
}

// TODO: Consider adding retries to the k8s operations
func (d deploymentConfiguration) Do(deploymentName, secretName, namespace string) (RollbackDeploymentFunc, error) {
	deploymentInterface := d.kubernetesInterface.AppsV1().Deployments(namespace)

	deployment, err := deploymentInterface.Get(context.TODO(), deploymentName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if len(deployment.Spec.Template.Spec.Containers) < 1 {
		return nil, fmt.Errorf("no containers found in %s/%s deployment", namespace, deploymentName)
	}
	envs := deployment.Spec.Template.Spec.Containers[0].Env
	previousSecretNamespacedName := ""
	for _, env := range envs {
		if env.Name == ConfigurationSecretEnvName {
			previousSecretNamespacedName = env.Value
			env.Value = fmt.Sprintf("%s/%s", namespace, secretName)
			break
		}
	}
	if previousSecretNamespacedName == "" {
		return nil, fmt.Errorf("no %s environment variable found in %s/%s deployment", ConfigurationSecretEnvName, namespace, name)
	}
	deployment.Spec.Template.Spec.Containers[0].Env = envs

	if _, err = deploymentInterface.Update(context.TODO(), deployment, v1.UpdateOptions{}); err != nil {
		return nil, err
	}

	// TODO: Wait until the deployment is rolled out
	return d.newRollbackDeploymentFunc(deploymentName, namespace, previousSecretNamespacedName), nil
}

func (d deploymentConfiguration) newRollbackDeploymentFunc(name, namespace, previousSecretNamespacedName string) RollbackDeploymentFunc {
	return func() error {
		deploymentInterface := d.kubernetesInterface.AppsV1().Deployments(namespace)

		deployment, err := deploymentInterface.Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return err
		}

		if len(deployment.Spec.Template.Spec.Containers) < 1 {
			return fmt.Errorf("no containers found in %s/%s deployment", namespace, name)
		}
		envs := deployment.Spec.Template.Spec.Containers[0].Env
		foundEnv := false
		for _, env := range envs {
			if env.Name == ConfigurationSecretEnvName {
				foundEnv = true
				env.Value = previousSecretNamespacedName
				break
			}
		}
		if foundEnv == false {
			return fmt.Errorf("no %s environment variable found in %s/%s deployment", ConfigurationSecretEnvName, namespace, name)
		}

		deployment.Spec.Template.Spec.Containers[0].Env = envs
		if _, err = deploymentInterface.Update(context.TODO(), deployment, v1.UpdateOptions{}); err != nil {
			return err
		}

		return nil
	}
}

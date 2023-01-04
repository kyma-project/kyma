package registry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	UsernameSecretKeyName = "username"
	PasswordSecretKeyName = "password"
	URLSecretKeyName      = "registryAddress"
)

var (
	functionRuntimeLabels = map[string]string{
		v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
	}
)

func GetFunctionDeploymentList(config *rest.Config) (*appsv1.DeploymentList, error) {
	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes Client: %v", err)
	}
	return listFunctionDeployments(k8sClient)
}

func listFunctionDeployments(k8sClient client.Client) (*appsv1.DeploymentList, error) {
	matchingLabels := client.MatchingLabels(functionRuntimeLabels)
	listOpts := &client.ListOptions{}
	matchingLabels.ApplyToList(listOpts)

	deploymentList := &appsv1.DeploymentList{}

	if err := k8sClient.List(context.Background(), deploymentList, listOpts); err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}
	return deploymentList, nil
}

func GetFunctionImage(d appsv1.Deployment) (reference.NamedTagged, error) {
	for _, container := range d.Spec.Template.Spec.Containers {
		if container.Name == "function" {
			ref, err := reference.ParseNamed(container.Image)
			if err != nil {
				return nil, fmt.Errorf("failed to parse Function container image: %v", err)
			}
			taggedRef, ok := ref.(reference.NamedTagged)
			if !ok {
				return nil, fmt.Errorf("failed to cast Function image name")
			}
			return taggedRef, nil
		}
	}
	return nil, fmt.Errorf("failed to find Function image")
}

func ReadRegistryConfigSecretOrDie(resetConfig *rest.Config, secretName, namespace string) *RegistryClientOptions {
	logger := log.Default()
	k8sClient, err := client.New(resetConfig, client.Options{})
	if err != nil {
		logger.Fatalf("failed to create kubernetes client: %v", err)
	}

	o, err := getConfigSecretData(k8sClient, secretName, namespace)
	if err != nil {
		logger.Fatalf("failed to get registry config secret data: %v", err)
	}
	return o
}

func getConfigSecretData(k8sClient client.Client, secretName, namespace string) (*RegistryClientOptions, error) {
	s := corev1.Secret{}

	if err := k8sClient.Get(context.Background(),
		types.NamespacedName{
			Name:      secretName,
			Namespace: namespace,
		}, &s); err != nil {
		return nil, fmt.Errorf("failed to get secret [%s/%s]: %v", secretName, namespace, err)

	}

	keys := []string{UsernameSecretKeyName, PasswordSecretKeyName, URLSecretKeyName}
	for _, key := range keys {
		if _, ok := s.Data[key]; !ok {
			return nil, fmt.Errorf(fmt.Sprintf("can't find required key [%s] in registry config secret", key))
		}
	}
	return &RegistryClientOptions{
		Username: string(s.Data[UsernameSecretKeyName]),
		Password: string(s.Data[PasswordSecretKeyName]),
		URL:      string(s.Data[URLSecretKeyName]),
	}, nil
}

func FunctionFromImageName(image string) (function, namespace string, err error) {
	if strings.Contains(image, "/cache") {
		return "", "", errors.New("invalid image name: cache image names can't be used")
	}
	parts := strings.Split(image, "-")
	if len(parts) != 2 {
		return "", "", errors.New("invalid image name")
	}
	return parts[0], parts[1], nil
}

func IsFunctionUpdating(config *rest.Config, name, namespace string) (bool, error) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)
	k8sClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return false, fmt.Errorf("failed to create Kubernetes Client: %v", err)
	}
	function := &v1alpha2.Function{}
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, function); err != nil {
		return false, fmt.Errorf("failed to get function: %v", err)
	}
	return function.IsUpdating(), nil
}

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/distribution/reference"
	"github.com/kyma-project/kyma/components/function-controller/internal/registry"
	"github.com/vrischmann/envconfig"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	UsernameSecretKeyName = "username"
	PasswordSecretKeyName = "password"
	URLSecretKeyName      = "registryAddress"
)

var (
	functionRuntimeLabels = map[string]string{
		"serverless.kyma-project.io/managed-by": "function-controller",
	}

	mainLog = ctrlzap.New().WithName("internal registry gc")
)

type config struct {
	Namespace               string `envconfig:"default=kyma-system"`
	RegistryConfigSecreName string `envconfig:"default=serverless-registry-config-default"`
}

func main() {
	mainLog.Info("reading configuration")
	cfg := &config{}
	if err := envconfig.Init(cfg); err != nil {
		mainLog.Error(err, "while reading env variables")
		os.Exit(1)

	}
	restConfig := ctrl.GetConfigOrDie()
	registryConfig := ReadRegistryConfigSecretOrDie(restConfig, cfg)

	deploymentList, err := getFunctionDeploymentList(restConfig)
	if err != nil {
		mainLog.Error(err, "while listing Function Deployments")
		os.Exit(1)
	}

	functionImages := registry.NewTaggedImageList()
	for _, deployment := range deploymentList.Items {
		tagged, err := getFunctionImage(deployment)
		if err != nil {
			mainLog.Error(err, "while parsing deployment images")
			os.Exit(1)
		}
		imageName := reference.Path(tagged)
		imageTag := tagged.Tag()
		functionImages.AddImageWithTag(imageName, imageTag)
	}

	registryClient, err := registry.NewRegistryClient(context.Background(),
		registryConfig)
	if err != nil {
		mainLog.Error(err, "while creating registry client")
		os.Exit(1)
	}

	for _, functionImage := range functionImages.ListImages() {
		repoCli, err := registryClient.ImageRepository(functionImage)
		if err != nil {
			mainLog.Error(err, "while creating repository client")
			os.Exit(1)
		}
		registryTags, err := repoCli.ListTags()
		if err != nil {
			mainLog.Error(err, "while getting image tags")
			os.Exit(1)
		}
		for _, tagStr := range registryTags {
			if !functionImages.HasImageWithTag(functionImage, tagStr) {
				tag, err := repoCli.GetImageTag(tagStr)
				if err != nil {
					mainLog.Error(err, "while getting tag details")
					os.Exit(1)
				}

				mainLog.Info(fmt.Sprintf("deleting image [%v:%v] with digest [%v]..", functionImage, tagStr, tag.Digest))
				err = repoCli.DeleteImageTag(tag.Digest)
				if err != nil {
					mainLog.Error(err, "while deleting image tag")
				}
				mainLog.Info(fmt.Sprintf("image [%v:%v] deleted successfully", functionImage, tagStr))
			}
		}

	}

}

func getFunctionDeploymentList(config *rest.Config) (*appsv1.DeploymentList, error) {
	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes Client: %v", err)
	}
	matchingLabels := client.MatchingLabels(functionRuntimeLabels)
	listOpts := &client.ListOptions{}
	matchingLabels.ApplyToList(listOpts)

	deploymentList := &appsv1.DeploymentList{}

	if err := k8sClient.List(context.Background(), deploymentList, listOpts); err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}
	return deploymentList, nil
}

func getFunctionImage(d appsv1.Deployment) (reference.NamedTagged, error) {
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

func ReadRegistryConfigSecretOrDie(resetConfig *rest.Config, envConfig *config) *registry.RegistryClientOptions {
	k8sClient, err := client.New(resetConfig, client.Options{})
	if err != nil {
		mainLog.Error(err, "while creating Kubernetes client")
		os.Exit(1)
	}

	s := corev1.Secret{}

	if err := k8sClient.Get(context.Background(),
		types.NamespacedName{
			Name:      envConfig.RegistryConfigSecreName,
			Namespace: envConfig.Namespace,
		}, &s); err != nil {
		mainLog.Error(err, "while getting secret")
		os.Exit(1)
	}

	keys := []string{UsernameSecretKeyName, PasswordSecretKeyName, URLSecretKeyName}
	for _, key := range keys {
		if _, ok := s.Data[key]; !ok {
			mainLog.Error(fmt.Errorf("can't find required key [%s] in registry config secret", key), "")
			os.Exit(1)
		}
	}
	return &registry.RegistryClientOptions{
		Username: string(s.Data[UsernameSecretKeyName]),
		Password: string(s.Data[PasswordSecretKeyName]),
		URL:      string(s.Data[URLSecretKeyName]),
	}
}

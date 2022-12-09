package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/distribution/reference"
	"github.com/kyma-project/kyma/components/function-controller/internal/registry"
	"github.com/vrischmann/envconfig"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type config struct {
	Namespace               string `envconfig:"default=kyma-system"`
	RegistryConfigSecreName string `envconfig:"default=serverless-registry-config-default"`
}

func main() {
	mainLog := ctrlzap.New().WithName("internal registry gc")
	mainLog.Info("reading configuration")
	cfg := &config{}
	if err := envconfig.Init(cfg); err != nil {
		mainLog.Error(err, "while reading env variables")
		os.Exit(1)

	}
	restConfig := ctrl.GetConfigOrDie()
	registryConfig := registry.ReadRegistryConfigSecretOrDie(restConfig, cfg.RegistryConfigSecreName, cfg.Namespace)

	deploymentList, err := registry.GetFunctionDeploymentList(restConfig)
	if err != nil {
		mainLog.Error(err, "while listing Function Deployments")
		os.Exit(1)
	}

	functionImages := registry.NewTaggedImageList()
	for _, deployment := range deploymentList.Items {
		tagged, err := registry.GetFunctionImage(deployment)
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

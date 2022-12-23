package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/distribution/reference"
	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/registry"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/rest"
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

	restConfig, registryConfig := readConfigurationOrDie(mainLog)

	registryClient, err := registry.NewRegistryClient(context.Background(), registryConfig)
	if err != nil {
		mainLog.Error(err, "while creating registry client")
		os.Exit(1)
	}

	functionImages, err := listFunctionImages(restConfig)
	if err != nil {
		mainLog.Error(err, "while listing function images")
		os.Exit(1)
	}

	for _, functionImage := range functionImages.ListImages() {
		if err := deleteUnreferencedTag(registryClient, functionImages, mainLog, functionImage); err != nil {
			mainLog.Error(err, "while deleting unreferenced tag")
			os.Exit(1)
		}
	}
}

func readConfigurationOrDie(l logr.Logger) (*rest.Config, *registry.RegistryClientOptions) {
	cfg := &config{}
	if err := envconfig.Init(cfg); err != nil {
		l.Error(err, "while reading env variables")
		os.Exit(1)
	}
	restConfig := ctrl.GetConfigOrDie()
	registryConfig := registry.ReadRegistryConfigSecretOrDie(restConfig, cfg.RegistryConfigSecreName, cfg.Namespace)

	return restConfig, registryConfig
}

func listFunctionImages(restConfig *rest.Config) (*registry.ImageList, error) {
	deploymentList, err := registry.GetFunctionDeploymentList(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Function Deployments")

	}

	functionImages := registry.NewTaggedImageList()

	for _, deployment := range deploymentList.Items {
		tagged, err := registry.GetFunctionImage(deployment)
		if err != nil {
			return nil, errors.Wrap(err, "while parsing deployment images")
		}
		imageName := reference.Path(tagged)
		imageTag := tagged.Tag()
		functionImages.AddImageWithTag(imageName, imageTag)
	}

	return &functionImages, nil
}

func deleteUnreferencedTag(cli registry.RegistryClient, imageList *registry.ImageList, l logr.Logger, image string) error {
	repoCli, err := cli.ImageRepository(image)
	if err != nil {
		return errors.Wrap(err, "while creating repository client")
	}
	registryTags, err := repoCli.ListTags()
	if err != nil {
		return errors.Wrap(err, "while getting image tags")
	}
	for _, tagStr := range registryTags {
		if !imageList.HasImageWithTag(image, tagStr) {
			err = deleteImageTag(repoCli, l, image, tagStr)
			if err != nil {
				l.Error(err, "while deleting image tag")
			} else {
				l.Info(fmt.Sprintf("image [%v:%v] deleted successfully", image, tagStr))
			}
		}
	}
	return nil
}

func deleteImageTag(cli registry.RepositoryClient, l logr.Logger, image, tagStr string) error {
	tag, err := cli.GetImageTag(tagStr)
	if err != nil {
		return errors.Wrap(err, "while getting tag details")
	}

	l.Info(fmt.Sprintf("deleting image [%v:%v] with digest [%v]..", image, tagStr, tag.Digest))
	return cli.DeleteImageTag(tag.Digest)
}

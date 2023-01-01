package main

import (
	"context"
	"fmt"
	"os"
	"strings"

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

	registryClient, err := registry.NewRegistryClient(context.Background(), registryConfig, mainLog)
	if err != nil {
		mainLog.Error(err, "while creating registry client")
		os.Exit(1)
	}
	mainLog.Info("removing function images")

	functionImages, err := listRunningFunctionsImages(restConfig)
	if err != nil {
		mainLog.Error(err, "while listing function images")
		os.Exit(1)
	}

	for _, functionImage := range functionImages.ListKeys() {
		if err := deleteUnreferencedTags(registryClient, restConfig, functionImages, mainLog, functionImage); err != nil {
			mainLog.Error(err, "while deleting unreferenced tag")
			os.Exit(1)
		}
	}

	mainLog.Info("removing cached layers")
	// It's better to get the list of layers _after_ we have deleted the unreferenced tags in the previous step.
	// This way we know we don't have any unused images on the registry.
	imageLayers, err := registryClient.ListRegistryImagesLayers()
	if err != nil {
		mainLog.Error(err, "while getting registry image layers")
		os.Exit(1)

	}
	cachedLayers, err := registryClient.ListRegistryCachedLayers()
	if err != nil {
		mainLog.Error(err, "while getting registry cached layers")
		os.Exit(1)

	}
	for _, cachedLayer := range cachedLayers.ListKeys() {
		// this cached layer is used by one of existing images, which means it's
		// likely to be used by an updated versions of the same function
		if imageLayers.HasKey(cachedLayer) {
			continue
		}
		if err := deleteCacheTags(registryClient, mainLog, cachedLayers.GetKey(cachedLayer)); err != nil {
			mainLog.Error(err, "while deleting unreferenced cached tag")
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

func listRunningFunctionsImages(restConfig *rest.Config) (*registry.NestedSet, error) {
	deploymentList, err := registry.GetFunctionDeploymentList(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Function Deployments")

	}
	functionImages := registry.NewNestedSet()

	for _, deployment := range deploymentList.Items {
		tagged, err := registry.GetFunctionImage(deployment)
		if err != nil {
			return nil, errors.Wrap(err, "while parsing deployment images")
		}
		imageName := reference.Path(tagged)
		imageTag := tagged.Tag()
		functionImages.AddKeyWithValue(imageName, imageTag)
	}
	return &functionImages, nil
}

func deleteUnreferencedTags(cli registry.RegistryClient, config *rest.Config, imageList *registry.NestedSet, l logr.Logger, image string) error {
	functionName, namespace, err := registry.FunctionFromImageName(image)
	if err != nil {
		return errors.Wrap(err, "while resolving function name")
	}
	functionUpdating, err := registry.IsFunctionUpdating(config, functionName, namespace)
	if err != nil {
		return errors.Wrap(err, "while checking function update status")
	}
	// we skip the whole image if the related function is updating
	if functionUpdating {
		return nil
	}
	return deleteImageTags(cli, imageList, l, image)
}

func deleteCacheTags(cli registry.RegistryClient, l logr.Logger, cachedImages map[string]struct{}) error {
	for c := range cachedImages {
		parts := strings.Split(c, ":")
		if len(parts) != 2 {
			return errors.New("invalid cached image name")
		}
		imageName := parts[0]
		tagStr := parts[1]

		repoCli, err := cli.ImageRepository(imageName)
		if err != nil {
			return err
		}
		err = deleteImageTag(repoCli, l, imageName, tagStr)
		if err != nil {
			l.Error(err, "while deleting image tag")
		} else {
			l.Info(fmt.Sprintf("image [%v:%v] deleted successfully", imageName, tagStr))
		}
	}
	return nil
}

func deleteImageTags(cli registry.RegistryClient, imageList *registry.NestedSet, l logr.Logger, image string) error {
	repoCli, err := cli.ImageRepository(image)
	if err != nil {
		return errors.Wrap(err, "while creating repository client")
	}
	registryTags, err := repoCli.ListTags()
	if err != nil {
		return errors.Wrap(err, "while getting image tags")
	}
	for _, tagStr := range registryTags {
		if imageList.HasKeyWithValue(image, tagStr) {
			continue
		}
		err = deleteImageTag(repoCli, l, image, tagStr)
		if err != nil {
			l.Error(err, "while deleting image tag")
		} else {
			l.Info(fmt.Sprintf("image [%v:%v] deleted successfully", image, tagStr))
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

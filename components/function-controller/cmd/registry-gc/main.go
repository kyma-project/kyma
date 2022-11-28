package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/kyma-project/kyma/components/function-controller/internal/registry"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	functionRuntimeLabels = map[string]string{
		"serverless.kyma-project.io/managed-by": "function-controller",
	}
)

func main() {
	k8sClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{})
	if err != nil {
		panic(fmt.Sprintf("failed to create Kubernetes Client: %v", err))
	}
	matchingLabels := client.MatchingLabels(functionRuntimeLabels)
	listOpts := &client.ListOptions{}
	matchingLabels.ApplyToList(listOpts)

	deploymentList := appsv1.DeploymentList{}

	if err := k8sClient.List(context.Background(), &deploymentList, listOpts); err != nil {
		panic(fmt.Sprintf("failed to list deployments: %v", err))
	}

	functionImages := map[string][]string{}
	for _, deployment := range deploymentList.Items {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == "function" {
				ref, err := reference.ParseNamed(container.Image)
				if err != nil {
					panic(fmt.Sprintf("failed to parse container image: %v", err))
				}
				taggedRef, ok := ref.(reference.NamedTagged)
				if !ok {
					panic("failed to cast image name")
				}
				functionImages[reference.Path(taggedRef)] = append(functionImages[reference.Path(taggedRef)], taggedRef.Tag())
			}
		}
	}

	regClient, err := registry.NewRegistryClient()
	regClient.Logf = func(format string, args ...interface{}) {}

	if err != nil {
		panic(fmt.Sprintf("failed to create registry client: %v", err))
	}

	// repos, err := regClient.Repositories()
	// if err != nil {
	// 	panic(fmt.Sprintf("failed to list registry images: %v", err))
	// }

	// registryImages := map[string][]string{}

	for image, tags := range functionImages {
		logrus.Infof("-----------------------------functionImages %v: %v", image, tags)

		registryTags, err := regClient.Tags(image)
		if err != nil {
			panic(fmt.Sprintf("failed to get image tags from registry: %v", err))
		}
		logrus.Infof("-----------------------------registryTags %v: %v", image, registryTags)

		for _, registryTag := range registryTags {
		xxxx:
			for _, functionTag := range tags {
				if functionTag == registryTag {
					logrus.Infof("-----------------------------image %v:%v is currently used, skipping", image, registryTag)

					break xxxx
				}
			}
			// d, err := regClient.ManifestDigest(image, registryTag)
			// if err != nil {
			// 	panic(fmt.Sprintf("failed to get image tag digest from registry: %v", err))
			// }
			m, err := regClient.ManifestV2(image, registryTag)
			if err != nil {
				panic(fmt.Sprintf("failed to get image tag digest from registry: %v", err))
			}
			logrus.Infof("-----------------------------deleting: %v:%v", image, m.Target().Digest)
			err = regClient.DeleteManifest(image, m.Target().Digest)
			if err != nil {
				if strings.Contains(err.Error(), "status=404") {
					logrus.Infof("manifest for %s:%s is already deleted", image, registryTag)
					continue
				}
				panic(fmt.Sprintf("failed to get image tag digest from registry: %v", err))
			}
			logrus.Infof("deleted successfully")

		}

		// for _, tag := range tags {
		// 	for _, regsiregistryTag := range registryTags {

		// 		if tag == regsiregistryTag {
		// 			logrus.Infof("-----------------------------image %v:%v is currently used, skipping", image, tag)

		// 			continue
		// 		}
		// 		// d, err := regClient.ManifestDigest(image, tag)
		// 		// if err != nil {
		// 		// 	panic(fmt.Sprintf("failed to get image tag digest from registry: %v", err))
		// 		// }
		// 		// m, err := regClient.ManifestV2(image, tag)
		// 		// if err != nil {
		// 		// 	panic(fmt.Sprintf("failed to get image tag digest from registry: %v", err))
		// 		// }
		// 		logrus.Infof("-----------------------------deleting: %v:%v", image, tag)
		// 		// err = regClient.DeleteManifest(image, m.Target().Digest)
		// 		// if err != nil {
		// 		// 	if strings.Contains(err.Error(), "status=404") {
		// 		// 		logrus.Infof("manifest for %s:%s is already deleted", image, tag)
		// 		// 		continue
		// 		// 	}
		// 		// 	panic(fmt.Sprintf("failed to get image tag digest from registry: %v", err))
		// 		// }
		// 		// logrus.Infof("deleted successfully")
		// 	}
		// }
	}

	// logrus.Infof("===================================================== repos: %v", repos)
	// logrus.Infof("===================================================== repos: %v", registryImages)
}

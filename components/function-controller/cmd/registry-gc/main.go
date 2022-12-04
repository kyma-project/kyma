package main

import (
	"context"
	"fmt"

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
		tagged, err := getFunctionImage(deployment)
		if err != nil {
			panic(err)
		}
		if tagged != nil {
			functionImages[reference.Path(tagged)] = append(functionImages[reference.Path(tagged)], tagged.Tag())
		}
	}

	logrus.Infof("------------------------------------- functionImages: %v", functionImages)
	registryClient, err := registry.NewRegistryClient(context.Background(),
		registry.RepositoryClientOptions{
			URL:      "http://localhost:5000",
			Username: "d3BZTGRTVnNCQ3BvMkNUeTBWMHc=",
			Password: "dDVGRFF3bEhFeXR1ckxxcVN0dXBTclNoUWJ1RFk2UXN5YWVabnVLTA==",
		})
	if err != nil {
		panic(err)
	}
	for function, funcTags := range functionImages {
		repoCli, err := registryClient.ImageRepository(function)
		if err != nil {
			panic(err)
		}

		registryTags, err := repoCli.ListTags()
		if err != nil {
			panic(err)
		}
		for _, tagStr := range registryTags {
			found := false
			for _, funcTag := range funcTags {
				if funcTag == tagStr {
					found = true
				}
				if found {
					continue
				}
				tag, err := repoCli.GetImageTag(tagStr)
				if err != nil {
					panic(err)
				}
				// 	err = repoCli.DeleteImageTag(tag.Digest)
				// 	if err != nil {
				// 		panic(err)
				// 	}
				// }
				logrus.Infof(" should delete: %v:%v, with digest: %v err: %v", function, tagStr, tag.Digest, err)
			}
		}

	}

}

func getFunctionImage(d appsv1.Deployment) (reference.NamedTagged, error) {
	for _, container := range d.Spec.Template.Spec.Containers {
		if container.Name == "function" {
			ref, err := reference.ParseNamed(container.Image)
			if err != nil {
				return nil, fmt.Errorf("failed to parse container image: %v", err)
			}
			taggedRef, ok := ref.(reference.NamedTagged)
			if !ok {
				return nil, fmt.Errorf("failed to cast image name")
			}
			return taggedRef, nil
		}
	}
	return nil, nil
}

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/manifestprocessor"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	manager "github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/resourcemanager"
)

const testIDLength = 8
const manifestsDirectory = "manifests/"
const commonResourcesFile = "no_access_strategy.yaml"
const resourceSeparator = "---"

func TestApiGatewayIntegration(t *testing.T) {

	k8sClient := getDynamicClient()

	t.Run("expose service without access strategy (plain access)", func(t *testing.T) {
		t.Parallel()
		testID := generateTestID()

		// create common resources from files
		commonResources, err := manifestprocessor.ParseFromFileWithTemplate(commonResourcesFile, manifestsDirectory, "", testID)
		if err != nil {
			panic(err)
		}
		for _, commonResource := range commonResources {
			resourceSchema, ns, _ := getResourceSchemaAndNamespace(commonResource)
			manager.CreateResource(k8sClient, resourceSchema, ns, commonResource)
		}

		//for _, commonResource := range commonResources {
		//	resourceSchema, ns, name := getResourceSchemaAndNamespace(commonResource)
		//	manager.UpdateResource(k8sClient, resourceSchema, ns, name, commonResource) //TODO: wait for resource creation
		//}

		for _, commonResource := range commonResources {
			resourceSchema, ns, name := getResourceSchemaAndNamespace(commonResource)
			manager.DeleteResource(k8sClient, resourceSchema, ns, name) // TODO: move delete after test execution
		}
		// TODO: create api-rule

		// TODO: wait until rules propagate

		// TODO: test response from service

		fmt.Println("test finished")
	})
}

func getDynamicClient() dynamic.Interface {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return client
}

func generateTestID() string {
	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, testIDLength)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func getResourceSchemaAndNamespace(manifest unstructured.Unstructured) (schema.GroupVersionResource, string, string) {
	metadata := manifest.Object["metadata"].(map[string]interface{})
	apiVersion := strings.Split(fmt.Sprintf("%s", manifest.Object["apiVersion"]), "/")
	namespace := "default"
	if metadata["namespace"] != nil {
		namespace = fmt.Sprintf("%s", metadata["namespace"])
	}
	resourceName := fmt.Sprintf("%s", metadata["name"])
	resourceKind := fmt.Sprintf("%s", manifest.Object["kind"])
	if resourceKind == "Namespace" {
		namespace = ""
	}
	//TODO: Move this ^ somewhere else and make it clearer
	apiGroup, version := getGroupAndVersion(apiVersion)
	resourceSchema := schema.GroupVersionResource{Group: apiGroup, Version: version, Resource: pluralForm(resourceKind)}
	return resourceSchema, namespace, resourceName
}

func getGroupAndVersion(apiVersion []string) (apiGroup string, version string) {
	if len(apiVersion) > 1 {
		return apiVersion[0], apiVersion[1]
	}
	return "", apiVersion[0]
}

func pluralForm(name string) string {
	return fmt.Sprintf("%ss", strings.ToLower(name))
}

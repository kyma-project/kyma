package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/callretry"
	"golang.org/x/oauth2/clientcredentials"
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
const commonResourcesFile = "common.yaml"
const testNamespaceFile = "test-ns.yaml"
const resourceSeparator = "---"

func TestApiGatewayIntegration(t *testing.T) {

	cfg := clientcredentials.Config{
		ClientID:     "client_id",
		ClientSecret: "uYWimVHjnYi6",
		TokenURL:     "https://oauth2.kyma.local/oauth2/token",
		Scopes:       []string{"read"},
	}

	c := cfg.Client(context.Background())
	_ = callretry.NewSecurityValidator(c, nil)

	k8sClient := getDynamicClient()

	// create namespace for testing
	nsResource, err := manifestprocessor.ParseFromFile(testNamespaceFile, manifestsDirectory, resourceSeparator)
	if err != nil {
		panic(err)
	}
	nsResourceSchema, ns, _ := getResourceSchemaAndNamespace(nsResource[0])
	// TODO: should not fail if namespace doesn't exists
	manager.CreateResource(k8sClient, nsResourceSchema, ns, nsResource[0])

	t.Run("expose service without access strategy (plain access)", func(t *testing.T) {
		t.Parallel()
		testID := generateTestID()

		// create common resources from files
		commonResources, err := manifestprocessor.ParseFromFileWithTemplate(commonResourcesFile, manifestsDirectory, resourceSeparator, testID)
		if err != nil {
			panic(err)
		}
		createResources(k8sClient, commonResources)

		//for _, commonResource := range commonResources {
		//	resourceSchema, ns, name := getResourceSchemaAndNamespace(commonResource)
		//	manager.UpdateResource(k8sClient, resourceSchema, ns, name, commonResource) //TODO: wait for resource creation
		//}

		// TODO: create api-rule

		// TODO: wait until rules propagate

		// TODO: test response from service

		deleteResources(k8sClient, commonResources)

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

func createResources(k8sClient dynamic.Interface, resources []unstructured.Unstructured) {
	for _, resource := range resources {
		resourceSchema, ns, _ := getResourceSchemaAndNamespace(resource)
		manager.CreateResource(k8sClient, resourceSchema, ns, resource)
	}
}

func deleteResources(k8sClient dynamic.Interface, resources []unstructured.Unstructured) {
	for _, resource := range resources {
		resourceSchema, ns, name := getResourceSchemaAndNamespace(resource)
		manager.DeleteResource(k8sClient, resourceSchema, ns, name)
	}
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

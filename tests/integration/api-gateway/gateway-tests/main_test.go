package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/api"
	"golang.org/x/oauth2/clientcredentials"

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
const OauthClientSecretLength = 8
const OauthClientIDLength = 8
const manifestsDirectory = "manifests/"
const perScenarioCommonResourcesFile = "per-scenario-commons.yaml"
const globalCommonResourcesFile = "global-commons.yaml"
const noAccessStrategyApiruleFile = "no_access_strategy.yaml"
const oauthStrategyApiruleFile = "oauth-strategy.yaml"
const jwtAndOauthStrategyApiruleFile = "jwt-oauth-strategy.yaml"
const resourceSeparator = "---"

func TestApiGatewayIntegration(t *testing.T) {
	var hydraAddr string
	var user string
	var pwd string
	var reqTimeout int

	flag.StringVar(&hydraAddr, "hydra-address", "", "Hydra service address")
	flag.StringVar(&user, "user", "", "User login to fetch JWT token")
	flag.StringVar(&pwd, "password", "", "User password to fetch JWT token")
	flag.IntVar(&reqTimeout, "request-timeout", 5, "Delay (in seconds) after which requests to API fail")

	flag.Parse()

	oauthClientID := generateRandomString(OauthClientIDLength)
	oauthClientSecret := generateRandomString(OauthClientSecretLength)

	_ = clientcredentials.Config{
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		TokenURL:     "https://oauth2.kyma.local/oauth2/token",
		Scopes:       []string{"read"},
	}

	k8sClient := getDynamicClient()

	// create common resources for all scenarios
	globalCommonResources, err := manifestprocessor.ParseFromFileWithTemplate(globalCommonResourcesFile, manifestsDirectory, resourceSeparator, struct {
		OauthClientSecret string
		OauthClientID     string
	}{
		OauthClientSecret: base64.StdEncoding.EncodeToString([]byte(oauthClientSecret)),
		OauthClientID:     base64.StdEncoding.EncodeToString([]byte(oauthClientID)),
	})
	if err != nil {
		panic(err)
	}
	createResources(k8sClient, globalCommonResources...)
	// defer deleting namespace (it will also delete all remaining resources in that namespace)
	defer func() {
		time.Sleep(time.Second * 3)
		nsResourceSchema, ns, name := getResourceSchemaAndNamespace(globalCommonResources[0])
		manager.DeleteResource(k8sClient, nsResourceSchema, ns, name)
	}()
	t.Run("parallel tests", func(t *testing.T) {
		t.Run("expose service without access strategy (plain access)", func(t *testing.T) {
			t.Parallel()
			testID := generateRandomString(testIDLength)

			// create common resources from files
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(perScenarioCommonResourcesFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				panic(err)
			}
			createResources(k8sClient, commonResources...)

			// create api-rule from file
			noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				panic(err)
			}
			createResources(k8sClient, noAccessStrategyApiruleResource...)

			//for _, commonResource := range commonResources {
			//	resourceSchema, ns, name := getResourceSchemaAndNamespace(commonResource)
			//	manager.UpdateResource(k8sClient, resourceSchema, ns, name, commonResource)
			//}

			// TODO: test response from service - jakkab

			deleteResources(k8sClient, commonResources...)

			fmt.Println("test finished")
		})
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

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
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

func createResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) {
	for _, resource := range resources {
		resourceSchema, ns, _ := getResourceSchemaAndNamespace(resource)
		manager.CreateResource(k8sClient, resourceSchema, ns, resource)
	}
}

func deleteResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) {
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

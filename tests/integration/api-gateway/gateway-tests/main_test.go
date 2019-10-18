package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const testIDLength = 8
const manifestsDirectory = "manifests/"
const commonResourcesFile = "common.yaml"
const resourceSeparator = "---"

func TestApiGatewayIntegration(t *testing.T) {

	k8sClient := getDynamicClient()

	// load common resource file
	commonResourcesRaw := getManifestsFromFile(commonResourcesFile)

	t.Run("expose service without access strategy (plain access)", func(t *testing.T) {
		t.Parallel()
		testID := generateTestID()

		// create common resources
		commonResources := getCommonResourcesForTest(testID, commonResourcesRaw...)
		for _, commonResource := range commonResources {
			fmt.Println(commonResource)
			resourceSchema, ns, name := getResourceSchemaAndNamespace(commonResource)
			fmt.Println(resourceSchema)
			createResource(k8sClient, resourceSchema, ns, commonResource)
			deleteResource(k8sClient, resourceSchema, ns, name) // TODO: move delete after test execution
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

func getManifestsFromFile(fileName string) []string {
	data, err := ioutil.ReadFile(manifestsDirectory + fileName)
	if err != nil {
		panic(err)
	}
	return strings.Split(string(data), resourceSeparator)
}

func parseTemplateWithData(templateRaw string, data interface{}) string {
	tmpl, err := template.New("tmpl").Parse(templateRaw)
	if err != nil {
		panic(err)
	}
	var resource bytes.Buffer
	err = tmpl.Execute(&resource, data)
	if err != nil {
		panic(err)
	}
	return resource.String()
}

func parseManifest(input []byte) (*unstructured.Unstructured, error) {
	var middleware map[string]interface{}
	err := json.Unmarshal(input, &middleware)
	if err != nil {
		return nil, err
	}

	resource := &unstructured.Unstructured{
		Object: middleware,
	}
	return resource, nil
}

func getCommonResourcesForTest(testID string, commonResourcesRaw ...string) []unstructured.Unstructured {
	var commonResources []unstructured.Unstructured
	for _, commonResourceRaw := range commonResourcesRaw {
		commonResourceYAML := parseTemplateWithData(commonResourceRaw, struct{ TestID string }{TestID: testID})
		commonResourceJSON, err := yaml.YAMLToJSON([]byte(commonResourceYAML))
		if err != nil {
			panic(err)
		}
		commonResource, err := parseManifest(commonResourceJSON)
		if err != nil {
			panic(err)
		}
		commonResources = append(commonResources, *commonResource)
	}
	return commonResources
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

	var apiGroup, version string
	if len(apiVersion) > 1 {
		apiGroup = apiVersion[0]
		version = apiVersion[1]
	} else {
		apiGroup = ""
		version = apiVersion[0]
	}
	resourceSchema := schema.GroupVersionResource{Group: apiGroup, Version: version, Resource: plurarForm(resourceKind)}
	return resourceSchema, namespace, resourceName
}

func plurarForm(name string) string {
	return fmt.Sprintf("%ss", strings.ToLower(name))
}

func createResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, manifest unstructured.Unstructured) {
	fmt.Println("Creating resource...")
	result, err := client.Resource(resourceSchema).Namespace(namespace).Create(&manifest, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created resource %q.\n", result.GetName())
}

func deleteResource(client dynamic.Interface, resourceSchema schema.GroupVersionResource, namespace string, resourceName string) {
	fmt.Println("Deleting resource...")
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if err := client.Resource(resourceSchema).Namespace(namespace).Delete(resourceName, deleteOptions); err != nil {
		panic(err)
	}

	fmt.Printf("Deleted resource %q.\n", resourceName)
}

package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const testIDLength = 8
const manifestsDirectory = "manifests/"
const commonResourcesFile = "common.yaml"
const resourceSeparator = "---"

func TestApiGatewayIntegration(t *testing.T) {

	//k8sClient := getDynamicClient()

	// load common resource file
	commonResourcesRaw := getManifestsFromFile(commonResourcesFile)

	t.Run("expose service without access strategy (plain access)", func(t *testing.T) {
		t.Parallel()
		testID := generateTestID()

		// parse go template to add unique id
		for _, commonResourceRaw := range commonResourcesRaw {
			fmt.Println(parseTemplateWithData(commonResourceRaw, struct{ TestID string }{TestID: testID}))
			// TODO: save output to a slice with common manifests
		}

		// TODO: create common resources from manifests

		// TODO: create api-rule

		// TODO: wait until rules propagate

		// TODO: test response from service

		fmt.Println("test 1")
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

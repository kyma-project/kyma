package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type flags struct {
	count      int
	kubeconfig *string
}

type httpLogPipeline struct {
	Name         string
	Tag          string
	Host         string
	StorageLimit string
}

func main() {
	var args = flags{}
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		args.kubeconfig = &kubeconfig
	} else if home := homedir.HomeDir(); home != "" {
		args.kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		args.kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.IntVar(&args.count, "count", 1, "Number of log pipelines to deploy")
	flag.Parse()

	if err := run(args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}

}

func run(f flags) error {
	config, err := clientcmd.BuildConfigFromFlags("", *f.kubeconfig)
	if err != nil {
		return err
	}

	for i := 0; i < f.count; i++ {
		logPipelineYAML, err := renderHTTPLogPipeline(i)
		if err != nil {
			return err
		}
		if err := createLogPipeline(config, logPipelineYAML); err != nil {
			return err
		}
	}

	return nil
}

func renderHTTPLogPipeline(count int) ([]byte, error) {
	rendered := bytes.Buffer{}
	values := httpLogPipeline{
		Name:         fmt.Sprintf("http-%d", count),
		Tag:          randomTag(),
		Host:         "foo",
		StorageLimit: "5M",
	}
	httpTempl := template.Must(template.ParseFiles("./assets/http.yml"))
	err := httpTempl.Execute(&rendered, values)

	if err != nil {
		return nil, err
	}
	return rendered.Bytes(), nil
}

func randomTag() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
	tagLength := 5
	chars := make([]rune, tagLength)
	for i := range chars {
		chars[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(chars)
}

func createLogPipeline(config *rest.Config, logPipelineYAML []byte) error {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	var logPipeline map[string]interface{}
	if err := yaml.Unmarshal(logPipelineYAML, &logPipeline); err != nil {
		return err
	}

	gvr := schema.GroupVersionResource{Group: "telemetry.kyma-project.io", Version: "v1alpha1", Resource: "logpipelines"}
	obj := unstructured.Unstructured{Object: logPipeline}

	fmt.Println("Creating a log pipeline...")

	result, err := dynamicClient.Resource(gvr).Create(context.TODO(), &obj, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	fmt.Println("Created a log pipeline ", result.GetName())

	return nil
}

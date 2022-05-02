package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type flags struct {
	count      int
	kubeconfig *string
}

type logpipelineHttp struct {
	Name         string
	Tag          string
	Host         string
	StorageLimit string
}

func main() {
	var args = flags{}
	if home := homedir.HomeDir(); home != "" {
		args.kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		args.kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.IntVar(&args.count, "count", 1, "Number of log pipelines to deploy")
	flag.Parse()

	out := os.Stdout
	if err := run(out, args); err != nil {
		fmt.Fprintf(out, "Error: %v\n", err)
		os.Exit(2)
	}

}

func run(out io.Writer, f flags) error {
	config, err := clientcmd.BuildConfigFromFlags("", *f.kubeconfig)
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	logPipelineYAML, err := renderHttpLogPipeline(f.count)
	if err != nil {
		return err
	}

	var logPipeline map[string]interface{}
	if err := yaml.Unmarshal(logPipelineYAML, &logPipeline); err != nil {
		return err
	}

	gvr := schema.GroupVersionResource{Group: "telemetry.kyma-project.io", Version: "v1alpha1", Resource: "logpipelines"}
	obj := unstructured.Unstructured{Object: logPipeline}

	fmt.Println("Creating log pipeline...")
	result, err := dynamicClient.Resource(gvr).Create(context.TODO(), &obj, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Created log piepline %q.\n", result.GetName())

	return nil
}

func renderHttpLogPipeline(count int) ([]byte, error) {
	rendered := bytes.Buffer{}
	values := logpipelineHttp{
		Name:         "foo",
		Tag:          "foo",
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

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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	logPipelineGVR = schema.GroupVersionResource{Group: "telemetry.kyma-project.io", Version: "v1alpha1", Resource: "logpipelines"}
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

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	lokiPipelineYAML, err := os.ReadFile("./assets/loki.yml")
	if err := createLogPipeline(dynamicClient, lokiPipelineYAML); err != nil {
		return err
	}

	for i := 0; i < f.count; i++ {
		httpPipelineYAML, err := renderHTTPLogPipeline(i)
		if err != nil {
			return err
		}

		if err := createLogPipeline(dynamicClient, httpPipelineYAML); err != nil {
			return err
		}
	}

	portForwardToPrometheus(config)
	time.Sleep(1 * time.Second)
	queryCPU := `avg(
		node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{cluster="", namespace="kyma-system", container="fluent-bit"}
	  * on(namespace,pod)
		group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{cluster="", namespace="kyma-system", workload_type="daemonset", workload="telemetry-fluent-bit"}
	  ) by (workload, workload_type)`

	queryMemory := `avg(
		container_memory_working_set_bytes{job="kubelet", metrics_path="/metrics/cadvisor", cluster="", namespace="kyma-system", container="fluent-bit", image!=""}
	  * on(namespace,pod)
		group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{cluster="", namespace="kyma-system", workload_type="daemonset", workload="telemetry-fluent-bit"}
	) by (workload, workload_type)`

	t := time.Now()

	resultCPU, err := queryPrometheus(queryCPU, t)
	if err != nil {
		return err
	}
	fmt.Printf("CPU result: %.3f at time: %s\n", resultCPU.Value, resultCPU.Timestamp.Time().Format(time.RFC3339Nano))

	resultMemory, err := queryPrometheus(queryMemory, t)
	if err != nil {
		return err
	}
	memory := formatBytes(int64(resultMemory.Value))
	fmt.Printf("Memory result: %s at time: %s\n", memory, resultMemory.Timestamp.Time().Format(time.RFC3339Nano))

	return nil
}

func renderHTTPLogPipeline(count int) ([]byte, error) {
	rendered := bytes.Buffer{}
	values := httpLogPipeline{
		Name: fmt.Sprintf("http-%d", count),
		Tag:  randomTag(),
		Host: "mockserver.mockserver",
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

func createLogPipeline(dynamicClient dynamic.Interface, logPipelineYAML []byte) error {
	logPipeline, err := toUnstructured(logPipelineYAML)
	if err != nil {
		return err
	}

	logPipelineName, err := name(logPipeline)
	if err != nil {
		return err
	}

	fmt.Printf("Creating a log pipeline %s...\n", logPipelineName)
	if _, err := dynamicClient.Resource(logPipelineGVR).Create(context.TODO(), logPipeline, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Log pipeline %s already exists\n", logPipelineName)
			return nil
		}
		return err
	}

	fmt.Printf("Created a log pipeline %s\n", logPipelineName)

	if err := waitForLogPipeline(dynamicClient, logPipeline); err != nil {
		return err
	}

	return nil
}

func waitForLogPipeline(dynamicClient dynamic.Interface, logPipeline *unstructured.Unstructured) error {
	logPipelineName, err := name(logPipeline)
	if err != nil {
		return err
	}

	watch, err := dynamicClient.Resource(logPipelineGVR).Watch(context.TODO(),
		metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s", logPipelineName)})
	if err != nil {
		return err
	}

	for event := range watch.ResultChan() {
		running, err := hasRunningCondition(event.Object.(*unstructured.Unstructured))
		if err != nil {
			return err
		}

		if running {
			fmt.Printf("Log pipeline %s is running\n", logPipelineName)
			return nil
		}

		fmt.Printf("Log pipeline %s is not yet running. Waiting...\n", logPipelineName)
	}

	return nil
}

func formatBytes(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

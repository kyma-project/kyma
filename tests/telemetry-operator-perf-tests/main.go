package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"golang.org/x/sync/errgroup"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	logPipelineGVR  = schema.GroupVersionResource{Group: "telemetry.kyma-project.io", Version: "v1alpha1", Resource: "logpipelines"}
	defaultTimeout  = 15 * time.Minute
	defaultWaitTime = 5 * time.Minute
)

type flags struct {
	count          int
	kubeconfig     *string
	timeout        time.Duration
	host           string
	port           int
	uri            string
	unhealthyRatio float64
	waitTime       time.Duration
}

type httpLogPipeline struct {
	Name string
	Tag  string
	Host string
	Port int
	URI  string
}

func main() {
	var f = flags{}
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		f.kubeconfig = &kubeconfig
	} else if home := homedir.HomeDir(); home != "" {
		f.kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		f.kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.IntVar(&f.count, "count", 1, "Number of http log pipelines to deploy")
	flag.DurationVar(&f.timeout, "timeout", defaultTimeout, "Timeout")
	flag.StringVar(&f.host, "host", "", "Http host log pipelines will send logs to")
	flag.StringVar(&f.uri, "uri", "", " URI Path")
	flag.IntVar(&f.port, "port", 80, "Http port log pipelines will send logs to")
	flag.Float64Var(&f.unhealthyRatio, "unhealthy-ratio", 0, "Ratio of unhealthy to healthy http log pipelines (from 0 to 1, where 0 means all are healthy and 1 means all are unhealthy)")
	flag.DurationVar(&f.timeout, "wait", defaultWaitTime, "Time to wait for before collecting the metrics")
	flag.Parse()

	if err := run(&f); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}
}

func run(f *flags) error {
	if f.host == "" {
		return errors.New("flag not specified: host")
	}

	if f.unhealthyRatio < 0 || f.unhealthyRatio > 1 {
		return errors.New("flag invalid: unhealthyRatio (should be between 0 and 1)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()

	config, err := clientcmd.BuildConfigFromFlags("", *f.kubeconfig)
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	lokiPipelineYAML, err := os.ReadFile("./deploy/loki.yml")
	if err := createLogPipeline(ctx, dynamicClient, lokiPipelineYAML); err != nil {
		return err
	}

	g, gctx := errgroup.WithContext(ctx)

	unhealthyCount := int(f.unhealthyRatio * float64(f.count))
	for i := 0; i < f.count; i++ {
		current := i
		g.Go(func() error {
			var httpPipelineYAML []byte
			if current < unhealthyCount {
				httpPipelineYAML, err = renderHTTPLogPipeline(f.host, "/bad", f.port, current)
			} else {
				httpPipelineYAML, err = renderHTTPLogPipeline(f.host, "/good", f.port, current)
			}
			if err != nil {
				return err
			}

			if err := createLogPipeline(gctx, dynamicClient, httpPipelineYAML); err != nil {
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	fmt.Printf("Waiting for %s before collecting the metrics", f.waitTime)
	time.Sleep(f.waitTime)

	if err := collectMetrics(ctx, config); err != nil {
		return err
	}

	return nil
}

func renderHTTPLogPipeline(host, uri string, port, index int) ([]byte, error) {
	rendered := bytes.Buffer{}
	values := httpLogPipeline{
		Name: fmt.Sprintf("http-%d", index),
		Tag:  randomTag(),
		Host: host,
		Port: port,
		URI:  uri,
	}
	httpTempl := template.Must(template.ParseFiles("./deploy/http.yml"))
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
		//#nosec G404
		chars[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(chars)
}

func createLogPipeline(ctx context.Context, dynamicClient dynamic.Interface, logPipelineYAML []byte) error {
	logPipeline, err := toUnstructured(logPipelineYAML)
	if err != nil {
		return err
	}

	logPipelineName, err := name(logPipeline)
	if err != nil {
		return err
	}

	fmt.Printf("Creating a log pipeline %s...\n", logPipelineName)
	if _, err := dynamicClient.Resource(logPipelineGVR).Create(ctx, logPipeline, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Log pipeline %s already exists\n", logPipelineName)
			return nil
		}
		return err
	}

	fmt.Printf("Created a log pipeline %s\n", logPipelineName)

	if err := waitForLogPipeline(ctx, dynamicClient, logPipeline); err != nil {
		return err
	}

	return nil
}

func waitForLogPipeline(ctx context.Context, dynamicClient dynamic.Interface, logPipeline *unstructured.Unstructured) error {
	logPipelineName, err := name(logPipeline)
	if err != nil {
		return err
	}

	selector := metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s", logPipelineName)}
	watch, err := dynamicClient.Resource(logPipelineGVR).Watch(ctx, selector)
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

func collectMetrics(ctx context.Context, config *rest.Config) error {
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

	resultCPU, err := queryPrometheus(ctx, queryCPU, t)
	if err != nil {
		return err
	}
	fmt.Printf("CPU result: %.3f at time: %s\n", resultCPU.Value, resultCPU.Timestamp.Time().Format(time.RFC3339Nano))

	resultMemory, err := queryPrometheus(ctx, queryMemory, t)
	if err != nil {
		return err
	}
	memory := formatBytes(int64(resultMemory.Value))
	fmt.Printf("Memory result: %s at time: %s\n", memory, resultMemory.Timestamp.Time().Format(time.RFC3339Nano))

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

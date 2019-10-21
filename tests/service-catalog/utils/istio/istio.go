package istio

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func podExec(cfg *restclient.Config, namespace, name, container string, command []string, output io.Writer) error {
	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	req := k8sClient.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Namespace(namespace).
		Name(name).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdout:    true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: output,
	})
	return err
}

func filterLines(input io.Reader, expectedString string) {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), expectedString) {
			fmt.Println(scanner.Text())
		}
	}
}

func DumpConfig(t *testing.T, host, labelSelector string) {
	podName := os.Getenv("POD_NAME")

	cfg, err := restclient.InClusterConfig()
	require.NoError(t, err)

	k8sClient, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	podList, err := k8sClient.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, podList.Items)
	var brokerPodName string
	var brokerNamespace string
	for _, pod := range podList.Items {
		fmt.Printf("The tested broker %s/%s IP=%s\n", pod.Namespace, pod.Name, pod.Status.PodIP)
		brokerPodName = pod.Name
		brokerNamespace = pod.Namespace
	}

	dumpIstioClustersConfig(t, cfg, podName, host)

	dumpEnvoyRBACConfig(t, cfg, brokerNamespace, brokerPodName)
}

func dumpIstioClustersConfig(t *testing.T, cfg *restclient.Config, podName, host string) {
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		err := podExec(cfg, "kyma-system", podName, "istio-proxy", []string{"curl", "http://localhost:15000/clusters", "-s"}, w)
		assert.NoError(t, err)
	}()

	fmt.Printf("Istio-proxy clusters config for the host %s:\n", host)
	filterLines(r, host)
	fmt.Println()
}

func dumpEnvoyRBACConfig(t *testing.T, cfg *restclient.Config, namespace, name string) {
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		err := podExec(cfg, namespace, name, "istio-proxy", []string{"curl", "http://localhost:15000/config_dump", "-s"}, w)
		assert.NoError(t, err)
	}()

	fmt.Printf("Envoy RBAC filters for %s/%s:\n", namespace, name)
	scanner := bufio.NewScanner(r)

	// if counter is greater than zero - print to the output
	counter := -1
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "envoy.filters.http.rbac") {
			counter = 120
		}
		if strings.Contains(line, "\"mixer\"") {
			counter = 0
		}
		counter = counter - 1

		if counter > 0 {
			fmt.Println(line)
		}
	}
	fmt.Println()
}

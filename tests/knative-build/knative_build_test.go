package knative_build_acceptance

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/avast/retry-go"
	build_api "github.com/knative/build/pkg/apis/build/v1alpha1"
	build "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	cpuLimits = "50m"
)

var kubeConfig *rest.Config
var k8sClient *kubernetes.Clientset

func TestKnativeBuild_Acceptance(t *testing.T) {
	kubeConfig = loadKubeConfigOrDie()
	buildClient := build.NewForConfigOrDie(kubeConfig).Builds("knative-build")
	labels := make(map[string]string)
	labels["test-build-func"] = "test-build"
	build, err := buildClient.Create(&build_api.Build{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-build",
			Labels:    labels,
			Namespace: "knative-build",
		},
		Spec: build_api.BuildSpec{
			Steps: []corev1.Container{{
				Name:    "test-build",
				Image:   "alpine:3.8",
				Command: []string{"echo", "-n", "hello build"},
			}},
		},
	})

	if err != nil {
		t.Fatalf("Unable to create build: %v", err)
	}

	defer deleteBuild(buildClient, build)

	k8sClient = kubernetes.NewForConfigOrDie(kubeConfig)
	if err != nil {
		log.Fatalf("Unable to get access to kubernetes: %v", err)
	}

	err = retry.Do(func() error {
		log.Printf("Checking build logs")
		podList, err := k8sClient.CoreV1().Pods("knative-build").List(meta.ListOptions{LabelSelector: "test-build-func=test-build"})
		if err != nil {
			return err
		}

		log.Printf("Number of pods: %v", len(podList.Items))

		if len(podList.Items) != 1 {
			return fmt.Errorf("unexpected number of test pods: we have more than one test pod: %v", podList)
		}

		req := k8sClient.CoreV1().Pods("knative-build").GetLogs(podList.Items[0].Name, &corev1.PodLogOptions{Container: "build-step-test-build"})
		podLogs, err := req.Stream()
		if err != nil {
			return fmt.Errorf("Error fetching logs: %v", err)
		}
		defer podLogs.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			return fmt.Errorf("Error in copying information from podLogs to buf")
		}

		if buf.String() != "hello build" {
			return fmt.Errorf("unexpected response: '%s'", buf.String())
		}

		return nil
	}, retry.OnRetry(func(n uint, err error) {
		log.Printf("[%v] try failed: %s", n, err)
	}), retry.Attempts(15),
	)

	if err != nil {
		t.Fatalf("cannot get expected response from build: %s", err)
	}
}

func loadKubeConfigOrDie() *rest.Config {
	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Cannot create in-cluster config: %v", err)
		}
		return cfg
	}

	var err error
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatalf("Cannot read kubeconfig: %s", err)
	}
	return kubeConfig
}

func deleteBuild(buildClient build.BuildInterface, build *build_api.Build) {
	var deleteImmediately int64
	_ = buildClient.Delete(build.Name, &meta.DeleteOptions{
		GracePeriodSeconds: &deleteImmediately,
	})
}

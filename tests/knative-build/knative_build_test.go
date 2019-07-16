package knative_build_acceptance

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/avast/retry-go"
	build_api "github.com/knative/build/pkg/apis/build/v1alpha1"
	build "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	cpuLimits = "50m"
)

func TestKnativeBuild_Acceptance(t *testing.T) {
	kubeConfig := loadKubeConfigOrDie()
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

	err = retry.Do(func() error {
		log.Printf("Checking build logs")
		cmd := exec.Command("kubectl", "-n", "knative-build", "logs", "-l", "test-build-func=test-build", "-c", "build-step-test-build")
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("[Error] Checking build logs: %v", string(stdoutStderr))
			return err
		}
		if string(stdoutStderr) != "hello build" {
			return fmt.Errorf("unexpected response: '%s'", string(stdoutStderr))
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

func MustGetenv(t *testing.T, name string) string {
	env := os.Getenv(name)
	if env == "" {
		t.Fatalf("Missing '%s' variable", name)
	}
	return env
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

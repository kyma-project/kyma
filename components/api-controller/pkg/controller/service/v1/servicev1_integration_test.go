package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestIntegrationGet(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping in short mode.")
	}

	services, err := servicesFromDefaultConfig()
	if err != nil {
		t.Error(err)
	}

	svc, err2 := services.get("kubernetes", "default")
	if err2 != nil {
		t.Errorf("Unable to get service. Root cause: %v", err2)
	}

	t.Logf("Kubernetes service: %v\n", svc)

	if svc == nil {
		t.Errorf("Kubernetes servie not found")
	}
}

func TestIntegrationGet_ShouldReturnNotFound(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping in short mode.")
	}

	services, err := servicesFromDefaultConfig()
	if err != nil {
		t.Error(err)
	}

	svc, err2 := services.get("notfoundservice", "default")
	if err2 != nil {
		t.Errorf("Unable to get service. Root cause: %v", err2)
	}

	if svc != nil {
		t.Logf("Kubernetes service: %v\n", svc)
		t.Errorf("Should not found the service.")
	}
}

func servicesFromDefaultConfig() (*services, error) {

	kubeConfigLocation := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		return nil, fmt.Errorf("Unable to load kube config. Root cause: %v", err)
	}

	clientset := kubernetes.NewForConfigOrDie(kubeConfig)

	return &services{clientset}, nil
}

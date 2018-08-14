package v1alpha2

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/crd"
	apiExtensionsClient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/tools/clientcmd"
)

func TestCrd_ShouldCreateCrd(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping in short mode.")
		return
	}

	registrar, err := registrarFromDefaultConfig()
	if err != nil {
		t.Errorf("Error: %+v", err)
		return
	}

	registrar.Register(Crd("kyma.local"))
}

func registrarFromDefaultConfig() (*crd.Registrar, error) {

	kubeConfigLocation := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		return nil, fmt.Errorf("unable to load kube config. Root cause: %v", err)
	}

	apiExtensionsClientSet, err2 := apiExtensionsClient.NewForConfig(kubeConfig)
	if err2 != nil {
		return nil, fmt.Errorf("unable to create API extensions client. Root cause: %v", err2)
	}

	return crd.NewRegistrar(apiExtensionsClientSet), nil
}

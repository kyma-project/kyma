package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	istioNetworking "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	"k8s.io/client-go/tools/clientcmd"
)

func TestIntegrationCreateUpdateDeleteVirtualService(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping in short mode.")
	}

	virtualServiceCtrl, err := defaultConfig()

	if err != nil {
		t.Fatal(err.Error())
	}

	metaDto := meta.Dto{
		Name:      "fake-vsvc",
		Namespace: "default",
	}

	dto := &Dto{
		MetaDto:     metaDto,
		ServiceName: "kubernetes",
		ServicePort: 443,
		Hostname:    "test.com",
	}

	t.Logf("Creating VirtualService %+v", dto)

	_, createErr := virtualServiceCtrl.Create(dto)
	if createErr != nil {
		t.Errorf("Unable to create virtualService. Root cause : %v", createErr)
	}

	updatedDto := dto
	updatedDto.Hostname = "changed.com"

	deleteVsvc := func() {
		t.Logf("Deleting VirtualService")
		deleteErr := virtualServiceCtrl.Delete(updatedDto)
		if deleteErr != nil {
			t.Errorf("Unable to delete VirtualService. Details : %s", deleteErr.Error())
		}
	}
	defer deleteVsvc()

	t.Logf("Updating VirtualService %+v with resource %+v", dto, updatedDto)
	_, updateErr := virtualServiceCtrl.Update(dto, updatedDto)
	if updateErr != nil {
		t.Errorf("Unable to update virtualService. Root cause : %v", updateErr)
	}
}

func defaultConfig() (Interface, error) {

	kubeConfigLocation := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		return nil, fmt.Errorf("Unable to load kube config. Root cause: %v", err)
	}

	clientset := istioNetworking.NewForConfigOrDie(kubeConfig)

	return New(clientset, testingGateway), nil
}

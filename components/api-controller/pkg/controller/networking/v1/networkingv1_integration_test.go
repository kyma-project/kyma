package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	istioNetworking "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	k8s "k8s.io/client-go/kubernetes"
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

	createdResource, createErr := virtualServiceCtrl.Create(dto)
	if createErr != nil {
		t.Fatalf("Unable to create virtualService. Root cause : %v", createErr)
	}

	updatedDto := *dto
	updatedDto.Hostname = "changed.com"

	// in the real use case the management of resources statuses is done automatically after the resource is created
	// also the update DTO is always created from different event than create DTO, and contain all needed information
	// here for testing purposes we need to set it manually
	dto.Status.Resource = *createdResource

	deleteVsvc := func() {
		t.Logf("Deleting VirtualService")
		deleteErr := virtualServiceCtrl.Delete(&updatedDto)
		if deleteErr != nil {
			t.Errorf("Unable to delete VirtualService. Details : %s", deleteErr.Error())
		}
	}
	defer deleteVsvc()

	t.Logf("Updating VirtualService %+v with resource %+v", dto, updatedDto)
	updatedResource, updateErr := virtualServiceCtrl.Update(dto, &updatedDto)
	if updateErr != nil {
		t.Errorf("Unable to update virtualService. Root cause : %v", updateErr)
	}
	updatedDto.Status.Resource = *updatedResource
}

func defaultConfig() (Interface, error) {

	kubeConfigLocation := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		return nil, fmt.Errorf("Unable to load kube config. Root cause: %v", err)
	}

	networkingClientset := istioNetworking.NewForConfigOrDie(kubeConfig)

	k8sClientset := k8s.NewForConfigOrDie(kubeConfig)

	return New(networkingClientset, k8sClientset, testingGateway, defaultCorsConfig()), nil
}

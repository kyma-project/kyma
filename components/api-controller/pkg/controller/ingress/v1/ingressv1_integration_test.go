package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestIntegrationCreateUpdateDeleteIngress(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping in short mode.")
	}

	ingCtrl, err := defaultConfig()

	if err != nil {
		t.Fatal(err.Error())
	}

	metaDto := meta.Dto{
		Namespace: "default",
	}

	dto := &Dto{
		MetaDto:     metaDto,
		ServiceName: "kubernetes",
		ServicePort: 443,
		Hostname:    "test.com",
	}

	t.Logf("Creating Ingress %+v", dto)

	_, createErr := ingCtrl.Create(dto)
	if createErr != nil {
		t.Errorf("Unable to create ingress. Root cause : %v", createErr)
	}

	updatedDto := dto
	updatedDto.Hostname = "changed.com"

	_, updateErr := ingCtrl.Update(dto, updatedDto)
	if updateErr != nil {
		t.Errorf("Unable to update ingress. Root cause : %v", updateErr)
	}

	deleteIng := func() {
		t.Logf("DELETE Ingress")
		deleteErr := ingCtrl.Delete(updatedDto)
		if deleteErr != nil {
			t.Errorf("Unable to delete Ingress. Details : %s", deleteErr.Error())
		}
	}

	defer deleteIng()
}

func defaultConfig() (*ingress, error) {

	kubeConfigLocation := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		return nil, fmt.Errorf("Unable to load kube config. Root cause: %v", err)
	}

	cs := k8s.NewForConfigOrDie(kubeConfig)

	return &ingress{cs}, nil
}

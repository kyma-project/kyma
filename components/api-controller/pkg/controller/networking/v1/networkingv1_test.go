package v1

import (
	"testing"

	fake "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned/fake"

	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	k8sCore "k8s.io/api/core/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testingGateway = "gatewayname.namespace.svc.cluster.local"
)

func TestCreateVirtualService(t *testing.T) {

	dto := fakeDto()
	fakeClientset := fake.NewSimpleClientset()

	virtualServiceCtrl := New(fakeClientset, testingGateway)
	_, err := virtualServiceCtrl.Create(dto)

	if err != nil {
		t.Errorf("Error creating VirtualService. Details : %s", err.Error())
	}
}

func TestCreateVirtualServiceForExistingVirtualService(t *testing.T) {

	dto := fakeDto()
	virtualService := toVirtualService(dto, testingGateway)
	fakeClientset := fake.NewSimpleClientset(virtualService)

	virtualServiceCtrl := New(fakeClientset, testingGateway)
	_, err := virtualServiceCtrl.Create(dto)

	if err == nil {
		t.Error("Should not create VirtualService because it already exsists but it did!")
	}
}
func TestUpdateVirtualService(t *testing.T) {

	oldApi := fakeDto()
	virtualService := toVirtualService(oldApi, testingGateway)

	t.Run("service assigned to virtualService has changed so virtualService will be updated", func(t *testing.T) {

		newApi := fakeDto()
		newApi.ServiceName = "fake-service"

		fakeClientset := fake.NewSimpleClientset(virtualService)

		virtualServiceCtrl := New(fakeClientset, testingGateway)
		updatedResource, err := virtualServiceCtrl.Update(oldApi, newApi)

		if err != nil {
			t.Errorf("Error while updating VirtualService. Details : %s", err.Error())
		} else if updatedResource == nil {
			t.Errorf("Error while updating VirtualService. VirtualService be udpated (old name: '%s', new name: '%s')", oldApi.ServiceName, newApi.ServiceName)
		}
	})

	t.Run("port of assigned service has changed so virtualService resource will be updated", func(t *testing.T) {
		newApi := oldApi
		newApi.ServicePort = 80

		fakeClientset := fake.NewSimpleClientset(virtualService)

		virtualServiceCtrl := New(fakeClientset, testingGateway)
		_, err := virtualServiceCtrl.Update(oldApi, newApi)

		if err != nil {
			t.Errorf("Error while updating VirtualService. Details : %s", err.Error())
		}
	})

	t.Run("nothing has changed so virtualService should not be updated", func(t *testing.T) {
		newApi := oldApi

		fakeClientset := fake.NewSimpleClientset(virtualService)

		virtualServiceCtrl := New(fakeClientset, testingGateway)
		updatedResource, err := virtualServiceCtrl.Update(oldApi, newApi)

		if err != nil {
			t.Errorf("Error while updating VirtualService. Details : %s", err.Error())
		}
		if updatedResource.Version != oldApi.Status.Resource.Version {
			t.Error("Error while updating VirtualService. Should not update virtualService because nothing has changed.")
		}
	})
}

func TestDeleteVirtualService(t *testing.T) {

	dto := fakeDto()
	virtualService := toVirtualService(dto, testingGateway)

	t.Run("Should delete virtual service if exists and dto not empty", func(t *testing.T) {
		fakeClientset := fake.NewSimpleClientset(virtualService)
		virtualServiceCtrl := New(fakeClientset, testingGateway)
		err := virtualServiceCtrl.Delete(dto)

		if err != nil {
			t.Errorf("Error deleting VirtualService. Details : %s", err.Error())
		}
	})

	t.Run("Should not delete virtual service if doesn't exist", func(t *testing.T) {
		fakeClientset := fake.NewSimpleClientset()
		virtualServiceCtrl := New(fakeClientset, testingGateway)
		err := virtualServiceCtrl.Delete(dto)

		if err == nil {
			t.Errorf("No error while deleting not existing VirtualService.")
		}
	})

}

func fakeDto() *Dto {
	return &Dto{
		MetaDto: meta.Dto{
			Name:      "fake-vsvc",
			Namespace: "default",
		},
		ServiceName: "kubernetes",
		ServicePort: 443,
		Hostname:    "fakeHostname.fakeDomain.com",
	}
}

func fakeService(name string, port int32) k8sCore.Service {
	return k8sCore.Service{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name: name,
		},
		Spec: k8sCore.ServiceSpec{
			Ports: []k8sCore.ServicePort{{Port: port}},
		},
	}
}

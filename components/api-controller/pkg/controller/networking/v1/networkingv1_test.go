package v1

import (
	"testing"

	fake "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned/fake"

	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	k8sCore "k8s.io/api/core/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	k8sFake "k8s.io/client-go/kubernetes/fake"
)

const (
	testingGateway   = "gatewayname.namespace.svc.cluster.local"
	defaultNamespace = "default"
	customNamespace  = "testing"
)

func TestCreateVirtualService(t *testing.T) {
	k8sClientset := k8sFake.NewSimpleClientset(fakeNamespace(defaultNamespace))

	t.Run("Should create a virtualservice if name and hostname is unique", func(t *testing.T) {
		dto := fakeDto()
		virtualService := toVirtualService(dto, testingGateway)
		fakeClientset := fake.NewSimpleClientset(virtualService)

		dto.MetaDto.Name = "another-fake-vsvc"
		dto.Hostname = "anotherFakeHostname.fakeDomain.com"

		virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
		_, err := virtualServiceCtrl.Create(dto)

		if err != nil {
			t.Errorf("Error creating VirtualService. Details : %s", err.Error())
		}
	})

	t.Run("Should not create a virtualservice if the same virtualservice already exists", func(t *testing.T) {
		dto := fakeDto()
		virtualService := toVirtualService(dto, testingGateway)
		fakeClientset := fake.NewSimpleClientset(virtualService)

		virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
		_, err := virtualServiceCtrl.Create(dto)

		if err == nil {
			t.Error("Should not create VirtualService because it already exists but it did!")
		}
	})

	t.Run("Should not create a virtualservice if hostname is already used by other virtualservice", func(t *testing.T) {
		dto := fakeDto()
		virtualService := toVirtualService(dto, testingGateway)
		// UID needs to be set manually for testing purposes. It is used to uniquely identify resource (virtualService).
		// Normally it is assigned by kubernetes after the resource is created, but fake clientset doesn't create it.
		virtualService.UID = types.UID("12345")
		fakeClientset := fake.NewSimpleClientset(virtualService)

		dto.MetaDto.Name = "another-fake-vsvc"

		virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
		_, err := virtualServiceCtrl.Create(dto)

		if err == nil {
			t.Error("Should not create VirtualService because the hostname is already used by other virtualservice but it did!")
		}
	})
}

func TestUpdateVirtualService(t *testing.T) {

	k8sClientset := k8sFake.NewSimpleClientset(fakeNamespace(defaultNamespace), fakeNamespace(customNamespace))

	oldApi := fakeDto()
	virtualService := toVirtualService(oldApi, testingGateway)
	// UID needs to be set manually for testing purposes. It is used to uniquely identify resource (virtualService).
	// Normally it is assigned by kubernetes after the resource is created, but fake clientset doesn't create it.
	virtualService.UID = types.UID("12345")

	t.Run("service assigned to virtualService has changed so virtualService will be updated", func(t *testing.T) {

		newApi := *oldApi
		newApi.ServiceName = "fake-service"

		fakeClientset := fake.NewSimpleClientset(virtualService)

		// Status of oldApi needs to be set manually for the test purposes
		// Normally it is assigned by kubernetes after the resource is created, but fake clientset doesn't create it.
		oldApi.Status.Resource = *gatewayResourceFrom(virtualService)

		virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
		updatedResource, err := virtualServiceCtrl.Update(oldApi, &newApi)

		if err != nil {
			t.Errorf("Error while updating VirtualService. Details : %s", err.Error())
		} else if updatedResource == nil {
			t.Errorf("Error while updating VirtualService. VirtualService be udpated (old name: '%s', new name: '%s')", oldApi.ServiceName, newApi.ServiceName)
		}
	})

	t.Run("port of assigned service has changed so virtualService resource will be updated", func(t *testing.T) {
		newApi := *oldApi
		newApi.ServicePort = 80

		fakeClientset := fake.NewSimpleClientset(virtualService)

		// Status of oldApi needs to be set manually for the test purposes
		// Normally it is assigned by kubernetes after the resource is created, but fake clientset doesn't create it.
		oldApi.Status.Resource = *gatewayResourceFrom(virtualService)

		virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
		_, err := virtualServiceCtrl.Update(oldApi, &newApi)

		if err != nil {
			t.Errorf("Error while updating VirtualService. Details : %s", err.Error())
		}
	})

	t.Run("nothing has changed so virtualService should not be updated", func(t *testing.T) {
		newApi := *oldApi

		fakeClientset := fake.NewSimpleClientset(virtualService)

		// Status of oldApi needs to be set manually for the test purposes
		// Normally it is assigned by kubernetes after the resource is created, but fake clientset doesn't create it.
		oldApi.Status.Resource = *gatewayResourceFrom(virtualService)

		virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
		updatedResource, err := virtualServiceCtrl.Update(oldApi, &newApi)

		if err != nil {
			t.Errorf("Error while updating VirtualService. Details : %s", err.Error())
		}
		if updatedResource.Version != oldApi.Status.Resource.Version {
			t.Error("Error while updating VirtualService. Should not update virtualService because nothing has changed.")
		}
	})

	t.Run("should not update virtualService if hostname is already used by other virtualservice", func(t *testing.T) {
		virtualServiceWithWantedHostname := toVirtualService(oldApi, testingGateway)
		virtualServiceWithWantedHostname.Name = "fake-vsvc-with-wanted-hostname"
		virtualServiceWithWantedHostname.Spec.Hosts = []string{"wanted-hostname"}
		virtualServiceWithWantedHostname.UID = "09876" // UID must be different than the UID of previously created virtualservice

		t.Run("in the same namespace", func(t *testing.T) {
			t.Logf("virtualService: %v+\nvirtualServiceWithWantedHostname: %v+\n", virtualService, virtualServiceWithWantedHostname)
			fakeClientset := fake.NewSimpleClientset(virtualService, virtualServiceWithWantedHostname)

			newApi := *oldApi
			newApi.Hostname = "wanted-hostname"

			// Status of oldApi needs to be set manually for the test purposes
			// Normally it is assigned by kubernetes after the resource is created, but fake clientset doesn't create it.
			oldApi.Status.Resource = *gatewayResourceFrom(virtualService)

			virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
			_, err := virtualServiceCtrl.Update(oldApi, &newApi)

			if err == nil {
				t.Errorf("Error did not occured while updating VirtualService, but should because hostname is already used by other virtualservice.")
			}
		})

		t.Run("in different namespace", func(t *testing.T) {
			virtualServiceWithWantedHostname.Namespace = customNamespace

			fakeClientset := fake.NewSimpleClientset(virtualService, virtualServiceWithWantedHostname)

			newApi := *oldApi
			newApi.Hostname = "wanted-hostname"

			// Status of oldApi needs to be set manually for the test purposes
			// Normally it is assigned by kubernetes after the resource is created, but fake clientset doesn't create it.
			oldApi.Status.Resource = *gatewayResourceFrom(virtualService)

			virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
			updatedResource, err := virtualServiceCtrl.Update(oldApi, &newApi)

			if err == nil {
				t.Errorf("Error did not occured while updating VirtualService, but should because hostname is already used by other virtualservice.")
			}
			t.Logf("%s", err)
			if updatedResource == nil {
				t.Error("Error while updating VirtualService. Should not delete previous virtualservice.")
			}
		})
	})

	t.Run("virtualService was not created due to already occupied hostname, so it should", func(t *testing.T) {
		virtualServiceWithWantedHostname := toVirtualService(oldApi, testingGateway)
		virtualServiceWithWantedHostname.Name = "fake-vsvc-with-wanted-hostname"
		virtualServiceWithWantedHostname.Spec.Hosts = []string{"wanted-hostname"}
		virtualServiceWithWantedHostname.UID = "09876" // UID must be different than the UID of previously created virtualservice

		t.Run("create the new VirtualService with valid and not occupied hostname", func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset()

			oldWrongApi := Dto{}
			newApi := fakeDto()

			virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
			updatedResource, err := virtualServiceCtrl.Update(&oldWrongApi, newApi)

			if err != nil {
				t.Errorf("Error while updating VirtualService, but should not because hostname is not used by other virtualservice. Details : %s", err.Error())
			}
			if updatedResource == nil {
				t.Error("Error while updating VirtualService. Should create virtualservice.")
			}
		})
		t.Run("not create the VirtualService if the hostname is occupied", func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset(virtualServiceWithWantedHostname)

			oldWrongApi := Dto{}
			newApi := fakeDto()
			newApi.Hostname = "wanted-hostname"

			virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
			updatedResource, err := virtualServiceCtrl.Update(&oldWrongApi, newApi)

			if err == nil {
				t.Errorf("Error did not occured while updating VirtualService, but should because hostname is used by other virtualservice.")
			}
			if updatedResource != nil {
				t.Error("Error while updating VirtualService. Should not create a virtualservice.")
			}
		})
	})
}

func TestDeleteVirtualService(t *testing.T) {

	k8sClientset := k8sFake.NewSimpleClientset(fakeNamespace(defaultNamespace))

	dto := fakeDto()
	virtualService := toVirtualService(dto, testingGateway)

	t.Run("Should delete virtual service if exists and dto not empty", func(t *testing.T) {
		fakeClientset := fake.NewSimpleClientset(virtualService)
		virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
		err := virtualServiceCtrl.Delete(dto)

		if err != nil {
			t.Errorf("Error deleting VirtualService. Details : %s", err.Error())
		}
	})

	t.Run("Should not return error if virtual service doesn't exist", func(t *testing.T) {
		fakeClientset := fake.NewSimpleClientset()
		virtualServiceCtrl := New(fakeClientset, k8sClientset, testingGateway)
		err := virtualServiceCtrl.Delete(dto)

		if err != nil {
			t.Errorf("Error while deleting non existing VirtualService. Details : %s", err.Error())
		}
	})
}

func fakeDto() *Dto {
	return &Dto{
		MetaDto: meta.Dto{
			Name:      "fake-vsvc",
			Namespace: defaultNamespace,
		},
		ServiceName: "kubernetes",
		ServicePort: 443,
		Hostname:    "fakeHostname.fakeDomain.com",
	}
}

func fakeNamespace(name string) *k8sCore.Namespace {
	return &k8sCore.Namespace{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name: name,
		},
	}
}

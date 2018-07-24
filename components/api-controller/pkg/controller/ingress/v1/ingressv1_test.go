package v1

import (
	"testing"

	k8sCore "k8s.io/api/core/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateIngress(t *testing.T) {

	dto := fakeDto()
	cs := fake.NewSimpleClientset()

	ingressCtrl := New(cs)
	_, err := ingressCtrl.Create(dto)

	if err != nil {
		t.Errorf("Error creating Ingress. Detials : %s", err.Error())
	}
}

func TestCreateIngForExistingIng(t *testing.T) {

	dto := fakeDto()
	ing := ingressFrom(dto)
	cs := fake.NewSimpleClientset(&ing)

	ingressCtrl := New(cs)
	_, err := ingressCtrl.Create(dto)

	if err == nil {
		t.Error("Should not create Ingress because it already exsists but it did!")
	}
}

func TestGetIngress(t *testing.T) {

	dto := fakeDto()
	ing := ingressFrom(dto)
	cs := fake.NewSimpleClientset(&ing)

	ingressCtrl := New(cs)
	_, err := ingressCtrl.Get(dto)

	if err != nil {
		t.Errorf("Error to get Ingress. Detials : %s", err.Error())
	}
}

func TestUpdateIngress(t *testing.T) {

	oldApi := fakeDto()
	ing := ingressFrom(oldApi)

	t.Run("service assigned to ingress has changed so ingress will be deleted and created again", func(t *testing.T) {

		svc := fakeService("fake-service", int32(oldApi.ServicePort))

		newApi := fakeDto()
		newApi.ServiceName = "fake-service"

		cs := fake.NewSimpleClientset(&ing, &svc)

		ingressCtrl := New(cs)
		updatedResource, err := ingressCtrl.Update(oldApi, newApi)

		if err != nil {
			t.Errorf("Error while updating Ingress. Details : %s", err.Error())
		} else if updatedResource == nil {
			t.Errorf("Error while updating Ingress. Ingress be udpated (old name: '%s', new name: '%s')", oldApi.ServiceName, newApi.ServiceName)
		} else if updatedResource.Name != newApi.ServiceName+"-ing" {
			t.Errorf("Error while updating Ingress. Ingress should have name : %s, but is: %s", newApi.ServiceName+"-ing", updatedResource.Name)
		}
	})

	t.Run("port of assigned service has changed so ingress resource will be updated", func(t *testing.T) {
		svc := fakeService(oldApi.ServiceName, 80)
		newApi := oldApi
		newApi.ServicePort = 80

		cs := fake.NewSimpleClientset(&ing, &svc)

		ingressCtrl := New(cs)
		_, err := ingressCtrl.Update(oldApi, newApi)

		if err != nil {
			t.Errorf("Error while updating Ingress. Details : %s", err.Error())
		}
	})

	t.Run("nothing has changed so ingress shouldn't be updated", func(t *testing.T) {
		newApi := oldApi

		cs := fake.NewSimpleClientset(&ing)

		ingressCtrl := New(cs)
		updatedResource, err := ingressCtrl.Update(oldApi, newApi)

		if err != nil {
			t.Errorf("Error while updating Ingress. Details : %s", err.Error())
		}
		if updatedResource != nil {
			t.Error("Error while updating Ingress. Should not update ingress because nothing has changed.")
		}
	})
}

func TestDeleteIngress(t *testing.T) {

	dto := fakeDto()
	ing := ingressFrom(dto)
	cs := fake.NewSimpleClientset(&ing)

	ingressCtrl := New(cs)
	err := ingressCtrl.Delete(dto)

	if err != nil {
		t.Errorf("Error deleting Ingress. Detials : %s", err.Error())
	}
}

func fakeDto() *Dto {
	return &Dto{
		ServiceName: "kubernetes",
		ServicePort: 443,
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

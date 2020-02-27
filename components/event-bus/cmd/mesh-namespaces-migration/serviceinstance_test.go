package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	servicecatalogv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogfakeclientset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
)

func TestNewserviceInstanceManager(t *testing.T) {
	testUserNamespaces := []string{
		"ns1",
		"ns2",
		// ns3 excluded
		"ns4",
	}

	testServiceInstances := []*servicecatalogv1beta1.ServiceInstance{
		// ns1
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns1-1",
				Namespace: "ns1",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns1-2",
				Namespace: "ns1",
			},
		},
		// ns2
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns2",
				Namespace: "ns2",
			},
		},
		// ns3
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns3",
				Namespace: "ns3",
			},
		},
	}

	cli := servicecatalogfakeclientset.NewSimpleClientset(
		serviceInstancesToObjectSlice(testServiceInstances)...,
	)

	m, err := newServiceInstanceManager(cli, testUserNamespaces)
	if err != nil {
		t.Fatalf("Failed to initialize serviceInstanceManager: %s", err)
	}

	// expect
	//  2 objects from ns1 (user namespace)
	//  1 object  from ns2 (user namespace)
	//  0 object  from ns3 (non-user namespace)
	//  0 object  from ns4 (does not contain any object)

	expect := sets.NewString(
		"ns1/my-events-ns1-1",
		"ns1/my-events-ns1-2",
		"ns2/my-events-ns2",
	)
	got := sets.NewString(
		serviceInstancesToKeys(m.serviceInstances)...,
	)

	if !got.Equal(expect) {
		t.Errorf("Unexpected ServiceInstances: (-:expect, +:got) %s", cmp.Diff(expect, got))
	}
}

func TestRecreateServiceInstance(t *testing.T) {
	testServiceInstance := &servicecatalogv1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-events",
			Namespace:       "ns",
			ResourceVersion: "00",
		},
		Spec: servicecatalogv1beta1.ServiceInstanceSpec{
			ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
				Name: "some-class",
			},
			ServicePlanRef: &servicecatalogv1beta1.LocalObjectReference{
				Name: "some-plan",
			},
		},
	}

	m := serviceInstanceManager{
		cli: servicecatalogfakeclientset.NewSimpleClientset(testServiceInstance),
	}

	err := m.recreateServiceInstance(*testServiceInstance)
	if err != nil {
		t.Fatalf("Failed to recreate ServiceInstance: %s", err)
	}

	svci, err := m.cli.ServicecatalogV1beta1().ServiceInstances("ns").Get("my-events", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Error getting ServiceInstance from cluster: %s", err)
	}

	if cmp.Diff(testServiceInstance, svci) == "" {
		t.Error("Expected new ServiceInstance to differ from original, got identical objects")
	}
}

func serviceInstancesToObjectSlice(svcis []*servicecatalogv1beta1.ServiceInstance) []runtime.Object {
	objects := make([]runtime.Object, len(svcis))
	for i := range svcis {
		objects[i] = svcis[i]
	}
	return objects
}

func serviceInstancesToKeys(svcis []servicecatalogv1beta1.ServiceInstance) []string {
	keys := make([]string, len(svcis))
	for i, svci := range svcis {
		keys[i] = fmt.Sprintf("%s/%s", svci.Namespace, svci.Name)
	}
	return keys
}

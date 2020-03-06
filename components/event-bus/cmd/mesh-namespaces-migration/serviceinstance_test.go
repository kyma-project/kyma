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

	testServiceClasses := []servicecatalogv1beta1.ServiceClass{
		// ns1
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-appbroker-class",
				Namespace: "ns1",
			},
			Spec: servicecatalogv1beta1.ServiceClassSpec{
				ServiceBrokerName: appBrokerServiceClass,
			},
		},
		// ns2
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-foo-class",
				Namespace: "ns2",
			},
			Spec: servicecatalogv1beta1.ServiceClassSpec{
				ServiceBrokerName: "foo-broker",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-appbroker-class",
				Namespace: "ns2",
			},
			Spec: servicecatalogv1beta1.ServiceClassSpec{
				ServiceBrokerName: appBrokerServiceClass,
			},
		},
		// ns4
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-helm-class",
				Namespace: "ns4",
			},
			Spec: servicecatalogv1beta1.ServiceClassSpec{
				ServiceBrokerName: "helm-broker",
			},
		},
	}

	testServiceInstances := []*servicecatalogv1beta1.ServiceInstance{
		// ns1
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns1-1",
				Namespace: "ns1",
			},
			Spec: servicecatalogv1beta1.ServiceInstanceSpec{
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-appbroker-class",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns1-2",
				Namespace: "ns1",
			},
			Spec: servicecatalogv1beta1.ServiceInstanceSpec{
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "does-no-exist",
				},
			},
		},
		// ns2
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns2-1",
				Namespace: "ns2",
			},
			Spec: servicecatalogv1beta1.ServiceInstanceSpec{
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-foo-class",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns2-2",
				Namespace: "ns2",
			},
			Spec: servicecatalogv1beta1.ServiceInstanceSpec{
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-appbroker-class",
				},
			},
		},
		// ns3
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns3",
				Namespace: "ns3",
			},
			Spec: servicecatalogv1beta1.ServiceInstanceSpec{
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-appbroker-class",
				},
			},
		},
		// ns4
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns4",
				Namespace: "ns4",
			},
			Spec: servicecatalogv1beta1.ServiceInstanceSpec{
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-helm-class",
				},
			},
		},
	}

	allObjects := append(
		serviceClassesToObjectSlice(testServiceClasses),
		serviceInstancesToObjectSlice(testServiceInstances)...,
	)

	cli := servicecatalogfakeclientset.NewSimpleClientset(allObjects...)

	m, err := newServiceInstanceManager(cli, testUserNamespaces)
	if err != nil {
		t.Fatalf("Failed to initialize serviceInstanceManager: %s", err)
	}

	// expect
	//  1 object from ns1 (user namespace, only 1 instance matching expected service class)
	//  1 object from ns2 (user namespace, only 1 instance matching expected service class)
	//  0 object from ns3 (non-user namespace)
	//  0 object from ns4 (does not contain a relevant service class)

	expect := sets.NewString(
		"ns1/my-events-ns1-1",
		"ns2/my-events-ns2-2",
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
		svcCatalogClient: servicecatalogfakeclientset.NewSimpleClientset(testServiceInstance),
	}

	err := m.recreateServiceInstance(*testServiceInstance)
	if err != nil {
		t.Fatalf("Failed to recreate ServiceInstance: %s", err)
	}

	svci, err := m.svcCatalogClient.ServicecatalogV1beta1().ServiceInstances("ns").Get("my-events", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Error getting ServiceInstance from cluster: %s", err)
	}

	if cmp.Diff(testServiceInstance, svci) == "" {
		t.Error("Expected new ServiceInstance to differ from original, got identical objects")
	}
}

func serviceClassesToObjectSlice(serviceClasses []servicecatalogv1beta1.ServiceClass) []runtime.Object {
	objects := make([]runtime.Object, len(serviceClasses))
	for i := range serviceClasses {
		objects[i] = &serviceClasses[i]
	}
	return objects
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

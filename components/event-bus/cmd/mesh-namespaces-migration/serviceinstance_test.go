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

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/applicationconnector/v1alpha1"
	kymaeventingfakeclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/fake"
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

	testEventActivations := []*applicationconnectorv1alpha1.EventActivation{
		// ns1
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-ea-ns1-1",
				Namespace: "ns1",
				OwnerReferences: []metav1.OwnerReference{
					// matches 2 ServiceInstances
					{
						APIVersion: "test/v1",
						Kind:       "Test",
						Name:       "dummy",
					},
					{
						APIVersion: "servicecatalog.k8s.io/v0",
						Kind:       serviceInstanceKind,
						Name:       "some-svci-ns1-1",
					},
					{
						APIVersion: "servicecatalog.k8s.io/v0",
						Kind:       serviceInstanceKind,
						Name:       "some-svci-ns1-2",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-ea-ns1-2",
				Namespace: "ns1",
				OwnerReferences: []metav1.OwnerReference{
					// matches 1 ServiceInstance
					{
						APIVersion: "servicecatalog.k8s.io/v0",
						Kind:       serviceInstanceKind,
						Name:       "some-svci-ns1-2",
					},
				},
			},
		},
		// ns2
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-ea-ns2",
				Namespace: "ns2",
				OwnerReferences: []metav1.OwnerReference{
					// matches 0 ServiceInstance
					{
						APIVersion: "test/v1",
						Kind:       "Test",
						Name:       "dummy",
					},
				},
			},
		},
		// ns3
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-ea-ns3",
				Namespace: "ns3",
				OwnerReferences: []metav1.OwnerReference{
					// matches 1 ServiceInstance
					{
						APIVersion: "servicecatalog.k8s.io/v0",
						Kind:       serviceInstanceKind,
						Name:       "some-svci-ns3",
					},
				},
			},
		},
		// ns4
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-ea-ns4-1",
				Namespace: "ns4",
				OwnerReferences: []metav1.OwnerReference{
					// matches 0 ServiceInstance
					{
						APIVersion: "test/v1",
						Kind:       "Test",
						Name:       "dummy",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-ea-ns4-2",
				Namespace: "ns4",
				OwnerReferences: []metav1.OwnerReference{
					// matches 1 ServiceInstance
					{
						APIVersion: "servicecatalog.k8s.io/v0",
						Kind:       serviceInstanceKind,
						Name:       "some-svci-ns4",
					},
				},
			},
		},
	}

	scObjects := append(
		serviceClassesToObjectSlice(testServiceClasses),
		serviceInstancesToObjectSlice(testServiceInstances)...,
	)
	scCli := servicecatalogfakeclientset.NewSimpleClientset(scObjects...)

	kymaCli := kymaeventingfakeclientset.NewSimpleClientset(
		eventActivationsToObjectSlice(testEventActivations)...,
	)

	m, err := newServiceInstanceManager(scCli, kymaCli, testUserNamespaces)
	if err != nil {
		t.Fatalf("Failed to initialize serviceInstanceManager: %s", err)
	}

	// expect
	//  1 ServiceInstance from ns1 (user namespace, only 1 instance matching expected service class)
	//  1 ServiceInstance from ns2 (user namespace, only 1 instance matching expected service class)
	//  0 ServiceInstance from ns3 (non-user namespace)
	//  0 ServiceInstance from ns4 (does not contain a relevant service class)

	expectSvci := sets.NewString(
		"ns1/my-events-ns1-1",
		"ns2/my-events-ns2-2",
	)
	gotSvci := sets.NewString(
		serviceInstancesToKeys(m.serviceInstances)...,
	)

	if !gotSvci.Equal(expectSvci) {
		t.Errorf("Unexpected ServiceInstances: (-:expect, +:got) %s", cmp.Diff(expectSvci, gotSvci))
	}

	// expect
	//  2 ServiceInstances from ns1 (multiple owner refs to different ServiceInstances)
	//  0 ServiceInstance  from ns2 (no matching owner ref)
	//  0 ServiceInstance  from ns3 (non-user namespace)
	//  1 ServiceInstance  from ns4 (single owner ref to single ServiceInstance)

	expectEA := eventActivationsByServiceInstanceAndNamespace{
		"ns1": eventActivationsByServiceInstance{
			"some-svci-ns1-1": []string{"my-ea-ns1-1"},
			"some-svci-ns1-2": []string{"my-ea-ns1-1", "my-ea-ns1-2"},
		},
		"ns4": eventActivationsByServiceInstance{
			"some-svci-ns4": []string{"my-ea-ns4-2"},
		},
	}
	gotEA := m.eventActivationIndex

	if diff := cmp.Diff(expectEA, gotEA); diff != "" {
		t.Errorf("Unexpected EventActivation index: (-:expect, +:got) %s", diff)
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

func eventActivationsToObjectSlice(eas []*applicationconnectorv1alpha1.EventActivation) []runtime.Object {
	objects := make([]runtime.Object, len(eas))
	for i := range eas {
		objects[i] = eas[i]
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

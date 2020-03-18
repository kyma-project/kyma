package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	dynamicfakeclientset "k8s.io/client-go/dynamic/fake"

	servicecatalogv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogfakeclientset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"

	appoperatorappconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	eventbusappconnectorv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/applicationconnector/v1alpha1"
	kymaeventingfakeclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/fake"
)

func TestNewserviceInstanceManager(t *testing.T) {
	testUserNamespaces := []string{
		"ns1",
		"ns2",
		// ns3 excluded
		"ns4",
	}

	testApplications := []*appoperatorappconnectorv1alpha1.Application{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-app-1",
			},
			Spec: appoperatorappconnectorv1alpha1.ApplicationSpec{
				Services: []appoperatorappconnectorv1alpha1.Service{
					// matches 2 "events" ServiceClasses IDs
					{
						ID: "my-appbroker-class-ns1-2",
						Entries: []appoperatorappconnectorv1alpha1.Entry{{
							Type: "Foo",
						}, {
							Type: eventsServiceEntryType,
						}, {
							Type: "Bar",
						}},
					},
					{
						ID: "my-appbroker-class-ns2",
						Entries: []appoperatorappconnectorv1alpha1.Entry{{
							Type: eventsServiceEntryType,
						}},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-app-2",
			},
			Spec: appoperatorappconnectorv1alpha1.ApplicationSpec{
				Services: []appoperatorappconnectorv1alpha1.Service{
					// matches 0 "events" ServiceClasses ID
					{
						ID: "my-appbroker-class-ns1-3",
						Entries: []appoperatorappconnectorv1alpha1.Entry{{
							Type: "Foo",
						}, {
							Type: "Bar",
						}},
					},
				},
			},
		},
	}

	testServiceClasses := []servicecatalogv1beta1.ServiceClass{
		// ns1
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-appbroker-class-ns1-1",
				Namespace: "ns1",
			},
			Spec: servicecatalogv1beta1.ServiceClassSpec{
				// has Application broker class,
				// is NOT referenced by any "events" Application
				ServiceBrokerName: appBrokerServiceClass,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-appbroker-class-ns1-2",
				Namespace: "ns1",
			},
			Spec: servicecatalogv1beta1.ServiceClassSpec{
				// has Application broker class,
				// is referenced by 1 "events" Application
				ServiceBrokerName: appBrokerServiceClass,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-appbroker-class-ns1-3",
				Namespace: "ns1",
			},
			Spec: servicecatalogv1beta1.ServiceClassSpec{
				// has Application broker class,
				// is NOT referenced by any "events" Application
				ServiceBrokerName: appBrokerServiceClass,
			},
		},
		// ns2
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-foo-class-ns2",
				Namespace: "ns2",
			},
			Spec: servicecatalogv1beta1.ServiceClassSpec{
				// has non- Application broker class
				ServiceBrokerName: "foo-broker",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-appbroker-class-ns2",
				Namespace: "ns2",
			},
			Spec: servicecatalogv1beta1.ServiceClassSpec{
				// has Application broker class
				ServiceBrokerName: appBrokerServiceClass,
			},
		},
		// ns4
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-helm-class-ns4",
				Namespace: "ns4",
			},
			Spec: servicecatalogv1beta1.ServiceClassSpec{
				// has non- Application broker class
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
				// references Application broker class
				// class is referenced by 1 "events" Application
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-appbroker-class-ns1-2",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns1-2",
				Namespace: "ns1",
			},
			Spec: servicecatalogv1beta1.ServiceInstanceSpec{
				// references non- Application broker class
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "does-no-exist",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns1-3",
				Namespace: "ns1",
			},
			Spec: servicecatalogv1beta1.ServiceInstanceSpec{
				// references Application broker class
				// class is referenced by a non- "events" Application
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-appbroker-class-ns1-3",
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
				// references non- Application broker class
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-foo-class-ns2",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-events-ns2-2",
				Namespace: "ns2",
			},
			Spec: servicecatalogv1beta1.ServiceInstanceSpec{
				// references Application broker class
				// class is referenced by 1 "events" Application
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-appbroker-class-ns2",
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
				// references Application broker class
				// class is NOT referenced by any "events" Application
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-appbroker-class-ns3",
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
				// references non- Application broker class
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-helm-class-ns4",
				},
			},
		},
	}

	testEventActivations := []*eventbusappconnectorv1alpha1.EventActivation{
		// ns1
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-ea-ns1-1",
				Namespace: "ns1",
				OwnerReferences: []metav1.OwnerReference{
					// references 2 ServiceInstances
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
					// references 1 ServiceInstance
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
					// references 0 ServiceInstance
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
					// references 1 ServiceInstance
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
					// references 0 ServiceInstance
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
					// references 1 ServiceInstance
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

	fakeScheme := runtime.NewScheme()
	if err := appoperatorappconnectorv1alpha1.AddToScheme(fakeScheme); err != nil {
		t.Fatalf("Failed to build fake Scheme: %s", err)
	}
	dynCli := dynamicfakeclientset.NewSimpleDynamicClient(fakeScheme,
		applicationsToObjectSlice(testApplications)...,
	)

	m, err := newServiceInstanceManager(scCli, kymaCli, dynCli, testUserNamespaces)
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
	gotEA := m.eventActivationsIndex

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

func applicationsToObjectSlice(apps []*appoperatorappconnectorv1alpha1.Application) []runtime.Object {
	objects := make([]runtime.Object, len(apps))
	for i, app := range apps {
		app.TypeMeta.APIVersion = appoperatorappconnectorv1alpha1.SchemeGroupVersion.String()
		app.TypeMeta.Kind = "Application"

		appData, _ := json.Marshal(app)
		appUnstr := &unstructured.Unstructured{}

		_ = json.Unmarshal(appData, appUnstr)

		objects[i] = appUnstr
	}
	return objects
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

func eventActivationsToObjectSlice(eas []*eventbusappconnectorv1alpha1.EventActivation) []runtime.Object {
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

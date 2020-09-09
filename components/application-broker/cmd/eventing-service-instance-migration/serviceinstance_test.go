package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	authenticationv1alpha1api "istio.io/api/authentication/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	dynamicfakeclientset "k8s.io/client-go/dynamic/fake"

	servicecatalogv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogfakeclientset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"

	appoperatorappconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

	istiov1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
	istiofakeclientset "istio.io/client-go/pkg/clientset/versioned/fake"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"

	appbrokerconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	kymaeventingfakeclientset "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
)

var _ k8sruntime.Object = (*istiov1alpha1.Policy)(nil)

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

	testEventActivations := []*appbrokerconnectorv1alpha1.EventActivation{
		// ns1
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "EventActivation",
				APIVersion: "applicationconnector.kyma-project.io",
			},
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
			Spec: appbrokerconnectorv1alpha1.EventActivationSpec{},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "EventActivation",
				APIVersion: "applicationconnector.kyma-project.io",
			},
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
			TypeMeta: metav1.TypeMeta{
				Kind:       "EventActivation",
				APIVersion: "applicationconnector.kyma-project.io",
			},
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
			TypeMeta: metav1.TypeMeta{
				Kind:       "EventActivation",
				APIVersion: "applicationconnector.kyma-project.io",
			},
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
			TypeMeta: metav1.TypeMeta{
				Kind:       "EventActivation",
				APIVersion: "applicationconnector.kyma-project.io",
			},
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
			TypeMeta: metav1.TypeMeta{
				Kind:       "EventActivation",
				APIVersion: "applicationconnector.kyma-project.io",
			},
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

	// This sample policy has been taken from a Kyma cluster
	// $k get policy -n eventmeshupgradetest eventmeshupgradetest-broker -oyaml
	//
	// apiVersion: authentication.istio.io/v1alpha1
	// kind: Policy
	// metadata:
	//   creationTimestamp: "2020-08-18T14:48:36Z"
	//   generation: 1
	//   labels:
	//     eventing.knative.dev/broker: default
	//   name: eventmeshupgradetest-broker
	//   namespace: eventmeshupgradetest
	//   resourceVersion: "39064"
	//   selfLink: /apis/authentication.istio.io/v1alpha1/namespaces/eventmeshupgradetest/policies/eventmeshupgradetest-broker
	//   uid: f616754c-ec77-4cae-9bac-b10e87292276
	// spec:
	//   peers:
	//   - mtls:
	//       mode: PERMISSIVE
	//   targets:
	//   - name: default-broker
	//   - name: default-broker-filter
	//
	testPolicies := []istiov1alpha1.Policy{
		// policy1
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Policy",
				APIVersion: "authentication.istio.io",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ns1-broker",
				Namespace: "ns1",
				Labels: map[string]string{
					policyKnativeBrokerLabelKey: policyKnativeBrokerLabelValue,
				},
			},
			Spec: authenticationv1alpha1api.Policy{
				Targets: []*authenticationv1alpha1api.TargetSelector{
					{
						Name: "default-broker",
					},
					{
						Name: "default-broker-filter",
					},
				},
				Peers: []*authenticationv1alpha1api.PeerAuthenticationMethod{{
					Params: &authenticationv1alpha1api.PeerAuthenticationMethod_Mtls{
						Mtls: &authenticationv1alpha1api.MutualTls{
							Mode: authenticationv1alpha1api.MutualTls_PERMISSIVE,
						}}},
				},
			},
		},
	}

	scObjects := append(
		serviceClassesToObjectSlice(testServiceClasses),
		serviceInstancesToObjectSlice(testServiceInstances)...,
	)

	fakeScheme := runtime.NewScheme()
	if err := appbrokerconnectorv1alpha1.AddToScheme(fakeScheme); err != nil {
		t.Fatalf("Failed to build fake Scheme: %s", err)
	}
	if err := appoperatorappconnectorv1alpha1.AddToScheme(fakeScheme); err != nil {
		t.Fatalf("Failed to build fake Scheme: %s", err)
	}
	if err := istiofakeclientset.AddToScheme(fakeScheme); err != nil {
		t.Fatalf("Failed to build fake Scheme: %s", err)
	}
	dynCli := dynamicfakeclientset.NewSimpleDynamicClient(fakeScheme,
		applicationsToObjectSlice(testApplications)...,
	)
	scCli := servicecatalogfakeclientset.NewSimpleClientset(scObjects...)

	kymaCli := kymaeventingfakeclientset.NewSimpleClientset(
		eventActivationsToObjectSlice(testEventActivations)...,
	)
	istioClient := istiofakeclientset.NewSimpleClientset(policiesToObjectSlice(testPolicies)...)
	// istioClient.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
	// 	resource := action.GetResource()
	// 	if action.GetVerb() == "DELETE" && resource.Group == "authentication.istio.io" {

	// 	}
	// 	// TODO:
	// 	return false, nil, nil
	// })
	m, err := newServiceInstanceManager(scCli, kymaCli.ApplicationconnectorV1alpha1(), dynCli, istioClient, testUserNamespaces)
	if err != nil {
		t.Fatalf("Failed to initialize serviceInstanceManager: %s", err)
	}
	if err := m.recreateAll(); err != nil {
		t.Fatalf("error while re-creating ServiceInstances: %v", err)
	}
	if err := m.deletePopulatedIstioPolicies(); err != nil {
		t.Fatalf("error while deleting Istio policies: %v", err)
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

	// ensure Istio Policies got deleted
	for _, policy := range testPolicies {
		policyDeleted := false
		for _, action := range istioClient.Actions() {
			if action.GetNamespace() == policy.Namespace && action.GetResource().Group == "authentication.istio.io" && action.GetVerb() == "delete" {
				policyDeleted = true
			}
		}
		if !policyDeleted {
			t.Errorf("No Istio Policy with name/namespace: %s/%s deleted!", policy.Name, policy.Namespace)
		}
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

func policiesToObjectSlice(policies []istiov1alpha1.Policy) []runtime.Object {
	objects := make([]runtime.Object, len(policies))
	for i := range policies {
		objects[i] = &policies[i]
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

func eventActivationsToObjectSlice(eas []*appbrokerconnectorv1alpha1.EventActivation) []runtime.Object {
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

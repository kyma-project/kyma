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
				// but it has no ServiceInstance
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
				// has Application broker class,
				// is referenced by 1 "events" Application
				// and it has 1 ServiceInstance
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
				// class is NOT referenced by "events" Application
				ServiceClassRef: &servicecatalogv1beta1.LocalObjectReference{
					Name: "my-appbroker-class-ns1-1",
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
				// with ServiceInstance
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

	testPolicies := []istiov1alpha1.Policy{
		// policy1
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Policy",
				APIVersion: "authentication.istio.io",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ns2-broker",
				Namespace: "ns2",
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

	istioClient := istiofakeclientset.NewSimpleClientset(policiesToObjectSlice(testPolicies)...)
	m, err := newServiceInstanceManager(scCli, dynCli, istioClient, testUserNamespaces)
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
	//  1 ServiceInstance from ns2 (user namespace, only 1 instance matching expected service class)

	expectSvci := sets.NewString(
		"ns2/my-events-ns2-2",
	)
	gotSvci := sets.NewString(
		serviceInstancesToKeys(m.serviceInstances)...,
	)

	if !gotSvci.Equal(expectSvci) {
		t.Errorf("unexpected ServiceInstances: (-:expect, +:got) %s", cmp.Diff(expectSvci, gotSvci))
	}

	// ensure Istio Policies got deleted
	for _, policy := range testPolicies {
		policies, err := istioClient.AuthenticationV1alpha1().Policies(policy.Namespace).List(metav1.ListOptions{})
		if err != nil {
			t.Fatalf("failed to list policies: %v", err)
		}
		if len(policies.Items) != 0 {
			t.Errorf("unexpected length of policies list found after migration:  (-:expect, +:got): -:%d +:%d ", 0, len(policies.Items))
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

func serviceInstancesToKeys(svcis []servicecatalogv1beta1.ServiceInstance) []string {
	keys := make([]string, len(svcis))
	for i, svci := range svcis {
		keys[i] = fmt.Sprintf("%s/%s", svci.Namespace, svci.Name)
	}
	return keys
}

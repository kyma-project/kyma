package controller_test

import (
	"context"
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	serviceCatalogInformers "github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultCacheSyncTimeout = time.Second * 5

func TestBindingLabelsFetcherHappyPath(t *testing.T) {

	type testCase struct {
		testName             string
		givenServiceInstance *v1beta1.ServiceInstance
		givenClass           runtime.Object
	}

	for _, tc := range []testCase{
		{
			testName:             "instance refers to external cluster service class name",
			givenServiceInstance: fixPromotionsServiceWithClusterServiceClassExternalName(),
			givenClass:           fixPromotionsClusterServiceClass(),
		},
		{
			testName:             "instance refers to direct cluster service class name",
			givenServiceInstance: fixPromotionsServiceWithDirectClusterServiceClassName(),
			givenClass:           fixPromotionsClusterServiceClass(),
		},
		{
			testName:             "instance refers to external service class name",
			givenServiceInstance: fixPromotionsServiceWithServiceClassExternalName(),
			givenClass:           fixPromotionsServiceClass(),
		},
		{
			testName:             "instance refers to direct service class name",
			givenServiceInstance: fixPromotionsServiceWithDirectServiceClassName(),
			givenClass:           fixPromotionsServiceClass(),
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			// GIVEN
			givenServiceBinding := fixPromotionsServiceBinding()
			fakeClientSet := fake.NewSimpleClientset(tc.givenServiceInstance, tc.givenClass)

			informerFactory := serviceCatalogInformers.NewSharedInformerFactory(fakeClientSet, time.Hour)

			sut := newLabelsFetcher(informerFactory)

			ctx, cancel := context.WithTimeout(context.Background(), defaultCacheSyncTimeout)
			defer cancel()

			informerFactory.Start(ctx.Done())
			informerFactory.WaitForCacheSync(ctx.Done())

			// WHEN
			actualLabels, err := sut.Fetch(givenServiceBinding)
			assert.NoError(t, err)
			expectedLabels := map[string]string{
				"access-label-1": "true",
			}
			// THEN
			assert.Equal(t, expectedLabels, actualLabels)
		})

	}
}

func TestBindingLabelsFetcherErrors(t *testing.T) {

	type testCase struct {
		testName         string
		givenScObjects   []runtime.Object
		givenBinding     *v1beta1.ServiceBinding
		expectedErrorMsg string
	}

	for _, tc := range []testCase{
		{
			testName:         "when binding not provided",
			expectedErrorMsg: "cannot fetch labels from ClusterServiceClass/ServiceClass because binding is nil",
		},
		{
			testName:     "when cannot find service instance pointed by ServiceBinding",
			givenBinding: fixPromotionsServiceBinding(),
			expectedErrorMsg: "while fetching ServiceInstance [promotions-service] from namespace [production] indicated by ServiceBinding:" +
				" serviceinstance.servicecatalog.k8s.io \"promotions-service\" not found",
		},
		{
			testName:       "when cannot find cluster service class",
			givenBinding:   fixPromotionsServiceBinding(),
			givenScObjects: []runtime.Object{fixPromotionsServiceWithClusterServiceClassExternalName()},
			expectedErrorMsg: "while fetching ClusterServiceClass [ac031e8c-9aa4-4cb7-8999-0d358726ffaa]:" +
				" clusterserviceclass.servicecatalog.k8s.io \"ac031e8c-9aa4-4cb7-8999-0d358726ffaa\" not found",
		},
		{
			testName:       "when cannot find service class",
			givenBinding:   fixPromotionsServiceBinding(),
			givenScObjects: []runtime.Object{fixPromotionsServiceWithServiceClassExternalName()},
			expectedErrorMsg: "while fetching ServiceClass [production/ac031e8c-9aa4-4cb7-8999-0d358726ffaa]: " +
				"serviceclass.servicecatalog.k8s.io \"ac031e8c-9aa4-4cb7-8999-0d358726ffaa\" not found",
		},
		{
			testName:       "when service instance refers to clusterserviceclass and serviceclass",
			givenBinding:   fixPromotionsServiceBinding(),
			givenScObjects: []runtime.Object{fixPromotionsServiceWithClusterServiceClassAndServiceClassReferences()},
			expectedErrorMsg: "unable to get class details because the ServiceInstance ServiceInstance production/promotions-service refers to " +
				"ClusterServiceClass ac031e8c-9aa4-4cb7-8999-0d358726ffab and ServiceClass ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			// GIVEN
			fakeClientSet := fake.NewSimpleClientset(tc.givenScObjects...)
			informerFactory := serviceCatalogInformers.NewSharedInformerFactory(fakeClientSet, time.Hour)
			sut := newLabelsFetcher(informerFactory)

			ctx, cancel := context.WithTimeout(context.Background(), defaultCacheSyncTimeout)
			defer cancel()

			informerFactory.Start(ctx.Done())
			informerFactory.WaitForCacheSync(ctx.Done())

			// WHEN
			_, err := sut.Fetch(tc.givenBinding)
			// THEN
			assert.Error(t, err)
			assert.EqualError(t, err, tc.expectedErrorMsg)
		})
	}
}

func fixPromotionsServiceBinding() *v1beta1.ServiceBinding {
	return &v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "promotions-service-binding",
			Namespace: "production",
		},
		Spec: v1beta1.ServiceBindingSpec{
			InstanceRef: v1beta1.LocalObjectReference{
				Name: "promotions-service",
			},
		},
	}
}

func fixPromotionsServiceClass() *v1beta1.ServiceClass {
	return &v1beta1.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
			Namespace: "production",
		},
		Spec: v1beta1.ServiceClassSpec{
			CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
				ExternalName: "promotions",
				ExternalMetadata: &runtime.RawExtension{
					Raw: []byte(`{"bindingLabels": {"access-label-1":"true"} }`),
				},
			},
		},
	}
}

func fixPromotionsClusterServiceClass() *v1beta1.ClusterServiceClass {
	return &v1beta1.ClusterServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
		},
		Spec: v1beta1.ClusterServiceClassSpec{
			CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
				ExternalName: "promotions",
				ExternalMetadata: &runtime.RawExtension{
					Raw: []byte(`{"bindingLabels": {"access-label-1":"true"} }`),
				},
			},
		},
	}
}

func fixPromotionsServiceWithClusterServiceClassExternalName() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "promotions-service",
			Namespace: "production",
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassExternalName: "promotions",
			},
			ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
				Name: "ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
			},
		},
	}
}

func fixPromotionsServiceWithDirectClusterServiceClassName() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "promotions-service",
			Namespace: "production",
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassName: "ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
			},
		},
	}
}

func fixPromotionsServiceWithServiceClassExternalName() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "promotions-service",
			Namespace: "production",
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassExternalName: "promotions",
			},
			ServiceClassRef: &v1beta1.LocalObjectReference{
				Name: "ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
			},
		},
	}
}

func fixPromotionsServiceWithDirectServiceClassName() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "promotions-service",
			Namespace: "production",
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassName: "ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
			},
		},
	}
}

func fixPromotionsServiceWithClusterServiceClassAndServiceClassReferences() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "promotions-service",
			Namespace: "production",
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassName:        "ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
				ClusterServiceClassName: "ac031e8c-9aa4-4cb7-8999-0d358726ffab",
			},
		},
	}
}

func newLabelsFetcher(informerFactory serviceCatalogInformers.SharedInformerFactory) *controller.BindingLabelsFetcher {
	return controller.NewBindingLabelsFetcher(informerFactory.Servicecatalog().V1beta1().ServiceInstances().Lister(),
		informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Lister(),
		informerFactory.Servicecatalog().V1beta1().ServiceClasses().Lister())
}

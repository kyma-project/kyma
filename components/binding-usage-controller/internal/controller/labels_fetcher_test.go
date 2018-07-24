package controller_test

import (
	"context"
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	serviceCatalogInformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

const defaultCacheSyncTimeout = time.Second * 5

func TestBindingLabelsFetcherHappyPath(t *testing.T) {

	type testCase struct {
		testName             string
		givenServiceInstance *v1beta1.ServiceInstance
	}

	for _, tc := range []testCase{
		{
			testName:             "instance refers to external class name",
			givenServiceInstance: fixPromotionsServiceWithClassExternalName(),
		},
		{
			testName:             "instance refers to direct class name",
			givenServiceInstance: fixPromotionsServiceWithDirectServiceClassName(),
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			// GIVEN
			givenServiceClass := fixPromotionsServiceClass()

			givenServiceBinding := fixPromotionsServiceBinding()
			fakeClientSet := fake.NewSimpleClientset(tc.givenServiceInstance, givenServiceClass)

			informerFactory := serviceCatalogInformers.NewSharedInformerFactory(fakeClientSet, time.Hour)

			sut := controller.NewBindingLabelsFetcher(informerFactory.Servicecatalog().V1beta1().ServiceInstances().Lister(), informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Lister())

			ctx, cancel := context.WithTimeout(context.Background(), defaultCacheSyncTimeout)
			defer cancel()

			informerFactory.Start(ctx.Done())

			cache.WaitForCacheSync(ctx.Done(), informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer().HasSynced, informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer().HasSynced)

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
			expectedErrorMsg: "cannot fetch labels from ClusterServiceClass because binding is nil",
		},
		{
			testName:     "when cannot find service instance pointed by ServiceBinding",
			givenBinding: fixPromotionsServiceBinding(),
			expectedErrorMsg: "while fetching ServiceInstance [promotions-service] from namespace [production] indicated by ServiceBinding:" +
				" serviceinstance.servicecatalog.k8s.io \"promotions-service\" not found",
		},
		{
			testName:       "when cannot find service class",
			givenBinding:   fixPromotionsServiceBinding(),
			givenScObjects: []runtime.Object{fixPromotionsServiceWithClassExternalName()},
			expectedErrorMsg: "while fetching ClusterServiceClass [ac031e8c-9aa4-4cb7-8999-0d358726ffaa]:" +
				" clusterserviceclass.servicecatalog.k8s.io \"ac031e8c-9aa4-4cb7-8999-0d358726ffaa\" not found",
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			// GIVEN
			fakeClientSet := fake.NewSimpleClientset(tc.givenScObjects...)
			informerFactory := serviceCatalogInformers.NewSharedInformerFactory(fakeClientSet, time.Hour)
			sut := controller.NewBindingLabelsFetcher(informerFactory.Servicecatalog().V1beta1().ServiceInstances().Lister(), informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Lister())

			ctx, cancel := context.WithTimeout(context.Background(), defaultCacheSyncTimeout)
			defer cancel()
			informerFactory.Start(ctx.Done())

			cache.WaitForCacheSync(ctx.Done(), informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer().HasSynced, informerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Informer().HasSynced)

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
			ServiceInstanceRef: v1beta1.LocalObjectReference{
				Name: "promotions-service",
			},
		},
	}
}

func fixPromotionsServiceClass() *v1beta1.ClusterServiceClass {
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

func fixPromotionsServiceWithClassExternalName() *v1beta1.ServiceInstance {
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

func fixPromotionsServiceWithDirectServiceClassName() *v1beta1.ServiceInstance {
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

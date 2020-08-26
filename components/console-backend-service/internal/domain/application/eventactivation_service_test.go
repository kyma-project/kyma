package application_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicFake "k8s.io/client-go/dynamic/fake"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestEventActivationService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		eventActivation1 := fixEventActivation("test", "test1")
		eventActivation2 := fixEventActivation("test", "test2")
		eventActivation3 := fixEventActivation("nope", "test3")

		dynamicClient, err := createDynamicClient(eventActivation1, eventActivation2, eventActivation3)
		require.NoError(t, err)
		informer := createFakeInformer(dynamicClient)

		svc := application.NewEventActivationService(informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		items, err := svc.List("test")

		require.NoError(t, err)
		assert.Len(t, items, 2)
		assert.ElementsMatch(t, items, []*v1alpha1.EventActivation{eventActivation1, eventActivation2})
	})

	t.Run("Not found", func(t *testing.T) {
		dynamicClient, err := createDynamicClient()
		require.NoError(t, err)
		informer := createFakeInformer(dynamicClient)
		svc := application.NewEventActivationService(informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		items, err := svc.List("test")

		require.NoError(t, err)
		assert.Len(t, items, 0)
	})
}

func fixEventActivation(namespace, name string) *v1alpha1.EventActivation {
	return &v1alpha1.EventActivation{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: v1.TypeMeta{
			Kind:       "EventActivation",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		Spec: v1alpha1.EventActivationSpec{
			SourceID:    "picco-bello",
			DisplayName: "aha!",
		},
	}
}

func createFakeInformer(dynamic dynamic.Interface) cache.SharedIndexInformer {
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamic, informerResyncPeriod)
	return informerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "eventactivations",
	}).Informer()
}

func createDynamicClient(objects ...runtime.Object) (*dynamicFake.FakeDynamicClient, error) {
	scheme := runtime.NewScheme()
	err := v1alpha1.AddToScheme(scheme)
	if err != nil {
		return &dynamicFake.FakeDynamicClient{}, err
	}
	result := make([]runtime.Object, len(objects))
	for i, obj := range objects {
		converted, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		result[i] = &unstructured.Unstructured{Object: converted}
	}
	return dynamicFake.NewSimpleDynamicClient(scheme, result...), nil
}

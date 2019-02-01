package application_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

func TestEventActivationService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		eventActivation1 := fixEventActivation("test", "test1")
		eventActivation2 := fixEventActivation("test", "test2")
		eventActivation3 := fixEventActivation("nope", "test3")

		informer := fixEventActivationInformer(eventActivation1, eventActivation2, eventActivation3)
		svc := application.NewEventActivationService(informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		items, err := svc.List("test")

		require.NoError(t, err)
		assert.Len(t, items, 2)
		assert.ElementsMatch(t, items, []*v1alpha1.EventActivation{eventActivation1, eventActivation2})
	})

	t.Run("Not found", func(t *testing.T) {
		informer := fixEventActivationInformer()
		svc := application.NewEventActivationService(informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		items, err := svc.List("test")

		require.NoError(t, err)
		assert.Len(t, items, 0)
	})
}

func fixEventActivation(environment, name string) *v1alpha1.EventActivation {
	return &v1alpha1.EventActivation{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: environment,
		},
		Spec: v1alpha1.EventActivationSpec{
			SourceID:    "picco-bello",
			DisplayName: "aha!",
		},
	}
}

func fixEventActivationInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	return informerFactory.Applicationconnector().V1alpha1().EventActivations().Informer()
}

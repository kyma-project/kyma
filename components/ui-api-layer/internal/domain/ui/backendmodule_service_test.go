package ui_test

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/ui"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/client/informers/externalversions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"testing"
	"time"
)

func TestBackendModuleService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		backendModule1 := fixBackendModule("1")
		backendModule2 := fixBackendModule("2")
		backendModule3 := fixBackendModule("3")

		backendModuleInformer := fixInformer(backendModule1, backendModule2, backendModule3)

		svc := ui.NewBackendModuleService(backendModuleInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, backendModuleInformer)

		instances, err := svc.List()
		require.NoError(t, err)
		assert.Equal(t, []*v1alpha1.BackendModule{
			backendModule1, backendModule2, backendModule3,
		}, instances)
	})

	t.Run("NotFound", func(t *testing.T) {
		backendModuleInformer := fixInformer()

		svc := ui.NewBackendModuleService(backendModuleInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, backendModuleInformer)

		var emptyArray []*v1alpha1.BackendModule
		instances, err := svc.List()
		require.NoError(t, err)
		assert.Equal(t, emptyArray, instances)
	})
}

func fixInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Ui().V1alpha1().BackendModules().Informer()

	return informer
}

func fixBackendModule(name string) *v1alpha1.BackendModule {
	return &v1alpha1.BackendModule{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	}
}

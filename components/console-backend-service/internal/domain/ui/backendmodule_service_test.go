package ui_test

import (
	"testing"
)

func TestBackendModuleService_List(t *testing.T) {
//	t.Run("Success", func(t *testing.T) {
//		backendModule1 := fixBackendModule("1")
//		backendModule2 := fixBackendModule("2")
//		backendModule3 := fixBackendModule("3")
//
//		backendModuleInformer := fixInformer(backendModule1, backendModule2, backendModule3)
//		svc := ui.NewBackendModuleService(backendModuleInformer)
//
//		testingUtils.WaitForInformerStartAtMost(t, time.Second, backendModuleInformer)
//
//		instances, err := svc.List()
//		require.NoError(t, err)
//		assert.Contains(t, instances, backendModule1)
//		assert.Contains(t, instances, backendModule2)
//		assert.Contains(t, instances, backendModule3)
//	})
//
//	t.Run("NotFound", func(t *testing.T) {
//		backendModuleInformer := fixInformer()
//
//		svc := ui.NewBackendModuleService(backendModuleInformer)
//
//		testingUtils.WaitForInformerStartAtMost(t, time.Second, backendModuleInformer)
//
//		var emptyArray []*v1alpha1.BackendModule
//		instances, err := svc.List()
//		require.NoError(t, err)
//		assert.Equal(t, emptyArray, instances)
//	})
//}
//
//func fixInformer(objects ...runtime.Object) cache.SharedIndexInformer {
//	client := fake.NewSimpleClientset(objects...)
//	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
//
//	informer := informerFactory.Ui().V1alpha1().BackendModules().Informer()
//
//	return informer
//}
//
//func fixBackendModule(name string) *v1alpha1.BackendModule {
//	module := v1alpha1.BackendModule{
//		ObjectMeta: v1.ObjectMeta{
//			Name: name,
//		},
//	}
//
//	return &module
}

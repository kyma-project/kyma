package module_test

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module/automock"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/client/informers/externalversions"
	"github.com/magiconair/properties/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	k8sTesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"testing"
	"time"
)

func TestMakePluggableFunc_PluggabilityDisabled(t *testing.T) {
	informer := fixInformer()
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	moduleMock := new(automock.PluggableModule)
	moduleMock.On("Enable").Return(nil).Once()
	defer moduleMock.AssertExpectations(t)

	makePluggable := module.MakePluggableFunc(informer, false)
	makePluggable(moduleMock)
}

func TestMakePluggableFunc_PluggabilityEnabled(t *testing.T) {
	t.Run("OnAdd-NoModuleWithGivenName", func(t *testing.T) {
		testName := "bar"
		obj := fixBackendModuleCR(testName)

		client := fake.NewSimpleClientset(obj)
		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		informer := informerFactory.Ui().V1alpha1().BackendModules().Informer()

		makePluggable := module.MakePluggableFunc(informer, true)

		fooModule := new(automock.PluggableModule)
		fooModule.On("Name").Return("foo")
		defer fooModule.AssertExpectations(t)
		makePluggable(fooModule)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	})

	t.Run("OnAdd-DisabledModule", func(t *testing.T) {
		testName := "bar"
		obj := fixBackendModuleCR(testName)

		client := fake.NewSimpleClientset(obj)
		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		informer := informerFactory.Ui().V1alpha1().BackendModules().Informer()

		makePluggable := module.MakePluggableFunc(informer, true)

		barModule := fixModuleMock("bar", false)
		barModule.On("Enable").Return(nil).Once()
		defer barModule.AssertExpectations(t)
		makePluggable(barModule)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	})

	t.Run("OnAdd-AlreadyEnabledModule", func(t *testing.T) {
		testName := "bar"
		obj := fixBackendModuleCR(testName)

		client := fake.NewSimpleClientset()
		//client.AddReactor("create", "backendmodules", func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
		//	log.Print("test-test")
		//	return true, obj, nil
		//})

		watcher := watch.NewFake()

		client.PrependWatchReactor("backendmodules", k8sTesting.DefaultWatchReactor(watcher, nil))

		// simulate add/update/delete watch events


		// test something that uses the client to do a watch
		informerFactory := externalversions.NewSharedInformerFactory(client, 0)
		informer := informerFactory.Ui().V1alpha1().BackendModules().Informer()
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		makePluggable := module.MakePluggableFunc(informer, true)

		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return("bar")
		moduleMock.On("IsEnabled").Return(true).Once()
		defer moduleMock.AssertExpectations(t)
		makePluggable(moduleMock)

		//_, err := client.UiV1alpha1().BackendModules("").Create(obj)
		//require.NoError(t, err)
		
		watcher.Add(obj)
		watcher.Delete(obj)

		assert.Equal(t, nil, informer.GetStore().List())
	})


}

func fixInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Ui().V1alpha1().BackendModules().Informer()

	return informer
}

func fixBackendModuleCR(name string) *v1alpha1.BackendModule {
	return &v1alpha1.BackendModule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func fixModuleMock(name string, isEnabled bool) *automock.PluggableModule {
	moduleMock := new(automock.PluggableModule)
	moduleMock.On("Name").Return(name)
	moduleMock.On("IsEnabled").Return(isEnabled).Once()

	return moduleMock
}
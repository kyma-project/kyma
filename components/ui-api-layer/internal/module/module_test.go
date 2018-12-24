package module_test

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module/automock"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/client/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	foo := "foo"
	bar := "bar"
	obj := fixBackendModuleCR(foo)

	client := fake.NewSimpleClientset(obj)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Ui().V1alpha1().BackendModules().Informer()

	makePluggable := module.MakePluggableFunc(informer, true)

	fooModule := new(automock.PluggableModule)
	fooModule.On("Name").Return(foo)
	fooModule.On("IsEnabled").Return(false).Once()
	fooModule.On("Enable").Return(nil).Once()
	defer fooModule.AssertExpectations(t)
	makePluggable(fooModule)

	barModule := new(automock.PluggableModule)
	barModule.On("Name").Return(bar)
	defer barModule.AssertExpectations(t)
	makePluggable(barModule)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
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

package module_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module/automock"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/client/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMakePluggableFunc_Pluggability(t *testing.T) {
	foo := "foo"
	bar := "bar"
	obj := fixBackendModuleCR(foo)

	client := fake.NewSimpleClientset(obj)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Ui().V1alpha1().BackendModules().Informer()

	makePluggable := module.MakePluggableFunc(informer)

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

func fixBackendModuleCR(name string) *v1alpha1.BackendModule {
	return &v1alpha1.BackendModule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

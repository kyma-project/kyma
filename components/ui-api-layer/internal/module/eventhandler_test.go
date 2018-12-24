package module

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestEventHandler_OnAdd(t *testing.T) {
	t.Run("NoModuleWithGivenName", func(t *testing.T) {
		obj := fixBackendModuleCR("foo")

		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return("bar")
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)
		eventHandler.OnAdd(obj)
	})

	t.Run("DisabledModule", func(t *testing.T) {
		name := "foo"
		obj := fixBackendModuleCR(name)

		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return(name)
		moduleMock.On("IsEnabled").Return(false)
		moduleMock.On("Enable").Return(nil).Once()
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)
		eventHandler.OnAdd(obj)
	})

	t.Run("AlreadyEnabledModule", func(t *testing.T) {
		name := "foo"
		obj := fixBackendModuleCR(name)

		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return(name)
		moduleMock.On("IsEnabled").Return(true)
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)
		eventHandler.OnAdd(obj)
	})
}

func TestEventHandler_OnUpdate(t *testing.T) {

}

func TestEventHandler_OnDelete(t *testing.T) {
	t.Run("NoModuleWithGivenName", func(t *testing.T) {
		obj := fixBackendModuleCR("foo")

		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return("bar")
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)
		eventHandler.OnDelete(obj)
	})

	t.Run("AlreadyDisabledModule", func(t *testing.T) {
		name := "foo"
		obj := fixBackendModuleCR(name)

		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return(name)
		moduleMock.On("IsEnabled").Return(false)
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)
		eventHandler.OnDelete(obj)
	})

	t.Run("EnabledModule", func(t *testing.T) {
		name := "foo"
		obj := fixBackendModuleCR(name)

		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return(name)
		moduleMock.On("IsEnabled").Return(true)
		moduleMock.On("Disable").Return(nil).Once()
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)
		eventHandler.OnDelete(obj)
	})
}

func fixBackendModuleCR(name string) *v1alpha1.BackendModule {
	return &v1alpha1.BackendModule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

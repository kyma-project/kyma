package module

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/module/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		moduleMock.On("IsEnabled").Return(false).Once()
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
		moduleMock.On("IsEnabled").Return(true).Once()
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)
		eventHandler.OnAdd(obj)
	})

	t.Run("Nil", func(t *testing.T) {
		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return("bar").Maybe()
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)

		require.NotPanics(t, func() {
			eventHandler.OnAdd(nil)
		})
	})
}

func TestEventHandler_OnUpdate(t *testing.T) {
	t.Run("NoModuleWithGivenName", func(t *testing.T) {
		oldObj := fixBackendModuleCR("foo")
		newObj := fixBackendModuleCR("foo2")

		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return("bar")
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)
		eventHandler.OnUpdate(oldObj, newObj)
	})

	t.Run("NothingToUpdate", func(t *testing.T) {
		oldObj := fixBackendModuleCR("foo")
		newObj := fixBackendModuleCR("foo")

		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return("foo").Maybe()
		moduleMock.On("IsEnabled").Return(true).Maybe()
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)
		eventHandler.OnUpdate(oldObj, newObj)
	})

	t.Run("EnableModule", func(t *testing.T) {
		oldObj := fixBackendModuleCR("foo")
		newObj := fixBackendModuleCR("bar")

		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return("bar")
		moduleMock.On("IsEnabled").Return(false).Once()
		moduleMock.On("Enable").Return(nil).Once()
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)
		eventHandler.OnUpdate(oldObj, newObj)
	})

	t.Run("DisableModule", func(t *testing.T) {
		oldObj := fixBackendModuleCR("foo")
		newObj := fixBackendModuleCR("bar")

		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return("foo")
		moduleMock.On("IsEnabled").Return(true).Once()
		moduleMock.On("Disable").Return(nil).Once()
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)
		eventHandler.OnUpdate(oldObj, newObj)
	})

	t.Run("Nil", func(t *testing.T) {
		obj := fixBackendModuleCR("foo")
		for _, tC := range []struct {
			old interface{}
			new interface{}
		}{
			{nil, nil},
			{obj, nil},
			{nil, obj},
		} {
			moduleMock := new(automock.PluggableModule)
			moduleMock.On("Name").Return("bar").Maybe()

			eventHandler := newEventHandler(moduleMock)

			require.NotPanics(t, func() {
				eventHandler.OnUpdate(tC.old, tC.new)
			})

			moduleMock.AssertExpectations(t)
		}
	})
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

	t.Run("Nil", func(t *testing.T) {
		moduleMock := new(automock.PluggableModule)
		moduleMock.On("Name").Return("bar").Maybe()
		defer moduleMock.AssertExpectations(t)

		eventHandler := newEventHandler(moduleMock)

		require.NotPanics(t, func() {
			eventHandler.OnDelete(nil)
		})
	})
}

func fixBackendModuleCR(name string) *v1alpha1.BackendModule {
	return &v1alpha1.BackendModule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

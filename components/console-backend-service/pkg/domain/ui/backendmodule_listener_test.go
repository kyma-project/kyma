package ui_test

import (
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/domain/ui"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/graph/model"
	"github.com/stretchr/testify/assert"
)

func TestPodListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		modelObj := new(model.BackendModule)
		k8sObj := new(v1alpha1.BackendModule)

		channel := make(chan *model.BackendModuleEvent, 1)
		listener := resource.NewListener(ui.NewBackendModuleEventHandler(channel))

		// when
		listener.OnAdd(k8sObj)
		result := <-channel

		// then
		assert.Equal(t, model.EventTypeAdd, result.Type)
		assert.Equal(t, *modelObj, result.Resource)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		listener := resource.NewListener(ui.NewBackendModuleEventHandler(nil))

		// when
		listener.OnAdd(new(v1alpha1.BackendModule))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		listener := resource.NewListener(ui.NewBackendModuleEventHandler(nil))

		// when
		listener.OnAdd(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		listener := resource.NewListener(ui.NewBackendModuleEventHandler(nil))

		// when
		listener.OnAdd(new(struct{}))
	})
}

func TestPodListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlPod := new(model.BackendModule)
		pod := new(v1alpha1.BackendModule)
		channel := make(chan *model.BackendModuleEvent, 1)

		listener := resource.NewListener(ui.NewBackendModuleEventHandler(channel))

		// when
		listener.OnDelete(pod)
		result := <-channel

		// then
		assert.Equal(t, model.EventTypeDelete, result.Type)
		assert.Equal(t, *gqlPod, result.Resource)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		listener := resource.NewListener(ui.NewBackendModuleEventHandler(nil))

		// when
		listener.OnDelete(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		listener := resource.NewListener(ui.NewBackendModuleEventHandler(nil))

		// when
		listener.OnDelete(new(struct{}))
	})
}

func TestPodListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlPod := new(model.BackendModule)
		pod := new(v1alpha1.BackendModule)

		channel := make(chan *model.BackendModuleEvent, 1)
		listener := resource.NewListener(ui.NewBackendModuleEventHandler(channel))

		// when
		listener.OnUpdate(pod, pod)
		result := <-channel

		// then
		assert.Equal(t, model.EventTypeUpdate, result.Type)
		assert.Equal(t, *gqlPod, result.Resource)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		listener := resource.NewListener(ui.NewBackendModuleEventHandler(nil))

		// when
		listener.OnUpdate(nil, nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		listener := resource.NewListener(ui.NewBackendModuleEventHandler(nil))

		// when
		listener.OnUpdate(new(struct{}), new(struct{}))
	})
}

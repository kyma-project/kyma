package ui_test

import (
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/apis/ui/v1alpha1"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/domain/ui"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
	"github.com/stretchr/testify/assert"
)

func TestPodListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlPod := new(model.BackendModule)
		module := new(v1alpha1.BackendModule)

		channel := make(chan *model.BackendModuleEvent, 1)
		podListener := ui.NewBackendModuleListener(channel, filterPodTrue)

		// when
		podListener.OnAdd(module)
		result := <-channel

		// then
		assert.Equal(t, model.EventTypeAdd, result.Type)
		assert.Equal(t, *gqlPod, result.Resource)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		podListener := ui.NewBackendModuleListener(nil, filterPodTrue)

		// when
		podListener.OnAdd(new(v1alpha1.BackendModule))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		podListener := ui.NewBackendModuleListener(nil, filterPodTrue)

		// when
		podListener.OnAdd(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		podListener := ui.NewBackendModuleListener(nil, filterPodTrue)

		// when
		podListener.OnAdd(new(struct{}))
	})
}

func TestPodListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlPod := new(model.BackendModule)
		pod := new(v1alpha1.BackendModule)
		channel := make(chan *model.BackendModuleEvent, 1)

		podListener := ui.NewBackendModuleListener(channel, filterPodTrue)

		// when
		podListener.OnDelete(pod)
		result := <-channel

		// then
		assert.Equal(t, model.EventTypeDelete, result.Type)
		assert.Equal(t, *gqlPod, result.Resource)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		podListener := ui.NewBackendModuleListener(nil, filterPodFalse)

		// when
		podListener.OnDelete(new(v1alpha1.BackendModule))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		podListener := ui.NewBackendModuleListener(nil, filterPodTrue)

		// when
		podListener.OnDelete(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		podListener := ui.NewBackendModuleListener(nil, filterPodTrue)

		// when
		podListener.OnDelete(new(struct{}))
	})
}

func TestPodListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlPod := new(model.BackendModule)
		pod := new(v1alpha1.BackendModule)

		channel := make(chan *model.BackendModuleEvent, 1)
		podListener := ui.NewBackendModuleListener(channel, filterPodTrue)

		// when
		podListener.OnUpdate(pod, pod)
		result := <-channel

		// then
		assert.Equal(t, model.EventTypeUpdate, result.Type)
		assert.Equal(t, *gqlPod, result.Resource)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		podListener := ui.NewBackendModuleListener(nil, filterPodFalse)

		// when
		podListener.OnUpdate(new(v1alpha1.BackendModule), new(v1alpha1.BackendModule))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		podListener := ui.NewBackendModuleListener(nil, filterPodTrue)

		// when
		podListener.OnUpdate(nil, nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		podListener := ui.NewBackendModuleListener(nil, filterPodTrue)

		// when
		podListener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterPodTrue(o *v1alpha1.BackendModule) bool {
	return true
}

func filterPodFalse(o *v1alpha1.BackendModule) bool {
	return false
}

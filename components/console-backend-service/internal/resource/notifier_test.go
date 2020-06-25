package resource

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource/automock"
)

func TestNotifier_AddListener(t *testing.T) {
	t.Run("Single", func(t *testing.T) {
		listener := new(automock.Listener)
		notifier := resource.NewNotifier()

		notifier.AddListener(listener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		listener := new(automock.Listener)
		notifier := resource.NewNotifier()

		notifier.AddListener(listener)
		notifier.AddListener(listener)
	})

	t.Run("Multiple", func(t *testing.T) {
		listenerA := new(automock.Listener)
		listenerB := new(automock.Listener)
		notifier := resource.NewNotifier()

		notifier.AddListener(listenerA)
		notifier.AddListener(listenerB)
	})

	t.Run("Nil listener", func(t *testing.T) {
		notifier := resource.NewNotifier()

		notifier.AddListener(nil)
	})
}

func TestNotifier_DeleteListener(t *testing.T) {
	t.Run("Single", func(t *testing.T) {
		listener := new(automock.Listener)
		notifier := resource.NewNotifier()

		notifier.AddListener(listener)
		notifier.DeleteListener(listener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		listener := new(automock.Listener)
		notifier := resource.NewNotifier()

		notifier.AddListener(listener)
		notifier.AddListener(listener)
		notifier.DeleteListener(listener)
	})

	t.Run("Multiple", func(t *testing.T) {
		listenerA := new(automock.Listener)
		listenerB := new(automock.Listener)
		notifier := resource.NewNotifier()

		notifier.AddListener(listenerA)
		notifier.AddListener(listenerB)
		notifier.DeleteListener(listenerA)
	})

	t.Run("Not existing", func(t *testing.T) {
		listener := new(automock.Listener)
		notifier := resource.NewNotifier()

		notifier.DeleteListener(listener)
	})

	t.Run("Nil listener", func(t *testing.T) {
		notifier := resource.NewNotifier()

		notifier.DeleteListener(nil)
	})
}

func TestNotifier_OnAdd(t *testing.T) {
	expected := new(struct{})

	t.Run("No listeners", func(t *testing.T) {
		notifier := resource.NewNotifier()

		notifier.OnAdd(expected)
	})

	t.Run("Single", func(t *testing.T) {
		listener := new(automock.Listener)
		listener.On("OnAdd", expected).Once()

		notifier := resource.NewNotifier()

		notifier.AddListener(listener)
		notifier.OnAdd(expected)

		listener.AssertCalled(t, "OnAdd", expected)
	})

	t.Run("Duplicated", func(t *testing.T) {
		listener := new(automock.Listener)
		listener.On("OnAdd", expected).Twice()

		notifier := resource.NewNotifier()

		notifier.AddListener(listener)
		notifier.AddListener(listener)
		notifier.OnAdd(expected)

		listener.AssertNumberOfCalls(t, "OnAdd", 2)
	})

	t.Run("Multiple", func(t *testing.T) {
		listenerA := new(automock.Listener)
		listenerB := new(automock.Listener)
		listenerA.On("OnAdd", expected).Once()
		listenerB.On("OnAdd", expected).Once()

		notifier := resource.NewNotifier()

		notifier.AddListener(listenerA)
		notifier.AddListener(listenerB)
		notifier.OnAdd(expected)

		listenerA.AssertCalled(t, "OnAdd", expected)
		listenerB.AssertCalled(t, "OnAdd", expected)
	})

	t.Run("Deleted listener", func(t *testing.T) {
		listenerA := new(automock.Listener)
		listenerB := new(automock.Listener)
		listenerA.On("OnAdd", expected).Once()
		listenerB.On("OnAdd", expected).Once()

		notifier := resource.NewNotifier()

		notifier.AddListener(listenerA)
		notifier.AddListener(listenerB)
		notifier.DeleteListener(listenerB)
		notifier.OnAdd(expected)

		listenerA.AssertCalled(t, "OnAdd", expected)
		listenerB.AssertNotCalled(t, "OnAdd", expected)
	})
}

func TestNotifier_OnDelete(t *testing.T) {
	expected := new(struct{})

	t.Run("No listeners", func(t *testing.T) {
		notifier := resource.NewNotifier()

		notifier.OnDelete(expected)
	})

	t.Run("Single", func(t *testing.T) {
		listener := new(automock.Listener)
		listener.On("OnDelete", expected).Once()

		notifier := resource.NewNotifier()

		notifier.AddListener(listener)
		notifier.OnDelete(expected)

		listener.AssertCalled(t, "OnDelete", expected)
	})

	t.Run("Duplicated", func(t *testing.T) {
		listener := new(automock.Listener)
		listener.On("OnDelete", expected).Twice()

		notifier := resource.NewNotifier()

		notifier.AddListener(listener)
		notifier.AddListener(listener)
		notifier.OnDelete(expected)

		listener.AssertNumberOfCalls(t, "OnDelete", 2)
	})

	t.Run("Multiple", func(t *testing.T) {
		listenerA := new(automock.Listener)
		listenerB := new(automock.Listener)
		listenerA.On("OnDelete", expected).Once()
		listenerB.On("OnDelete", expected).Once()

		notifier := resource.NewNotifier()

		notifier.AddListener(listenerA)
		notifier.AddListener(listenerB)
		notifier.OnDelete(expected)

		listenerA.AssertCalled(t, "OnDelete", expected)
		listenerB.AssertCalled(t, "OnDelete", expected)
	})

	t.Run("Deleted listener", func(t *testing.T) {
		listenerA := new(automock.Listener)
		listenerB := new(automock.Listener)
		listenerA.On("OnDelete", expected).Once()
		listenerB.On("OnDelete", expected).Once()

		notifier := resource.NewNotifier()

		notifier.AddListener(listenerA)
		notifier.AddListener(listenerB)
		notifier.DeleteListener(listenerB)
		notifier.OnDelete(expected)

		listenerA.AssertCalled(t, "OnDelete", expected)
		listenerB.AssertNotCalled(t, "OnDelete", expected)
	})
}

func TestNotifier_OnUpdate(t *testing.T) {
	expectedNew := new(struct{})
	expectedOld := new(struct{})

	t.Run("No listeners", func(t *testing.T) {
		notifier := resource.NewNotifier()

		notifier.OnUpdate(expectedOld, expectedNew)
	})

	t.Run("Single", func(t *testing.T) {
		listener := new(automock.Listener)
		listener.On("OnUpdate", expectedOld, expectedNew).Once()

		notifier := resource.NewNotifier()

		notifier.AddListener(listener)
		notifier.OnUpdate(expectedOld, expectedNew)

		listener.AssertCalled(t, "OnUpdate", expectedOld, expectedNew)
	})

	t.Run("Duplicated", func(t *testing.T) {
		listener := new(automock.Listener)
		listener.On("OnUpdate", expectedOld, expectedNew).Twice()

		notifier := resource.NewNotifier()

		notifier.AddListener(listener)
		notifier.AddListener(listener)
		notifier.OnUpdate(expectedOld, expectedNew)

		listener.AssertNumberOfCalls(t, "OnUpdate", 2)
	})

	t.Run("Multiple", func(t *testing.T) {
		listenerA := new(automock.Listener)
		listenerB := new(automock.Listener)
		listenerA.On("OnUpdate", expectedOld, expectedNew).Once()
		listenerB.On("OnUpdate", expectedOld, expectedNew).Once()

		notifier := resource.NewNotifier()

		notifier.AddListener(listenerA)
		notifier.AddListener(listenerB)
		notifier.OnUpdate(expectedOld, expectedNew)

		listenerA.AssertCalled(t, "OnUpdate", expectedOld, expectedNew)
		listenerB.AssertCalled(t, "OnUpdate", expectedOld, expectedNew)
	})

	t.Run("Deleted listener", func(t *testing.T) {
		listenerA := new(automock.Listener)
		listenerB := new(automock.Listener)
		listenerA.On("OnUpdate", expectedOld, expectedNew).Once()
		listenerB.On("OnUpdate", expectedOld, expectedNew).Once()

		notifier := resource.NewNotifier()

		notifier.AddListener(listenerA)
		notifier.AddListener(listenerB)
		notifier.DeleteListener(listenerB)
		notifier.OnUpdate(expectedOld, expectedNew)

		listenerA.AssertCalled(t, "OnUpdate", expectedOld, expectedNew)
		listenerB.AssertNotCalled(t, "OnUpdate", expectedOld, expectedNew)
	})
}

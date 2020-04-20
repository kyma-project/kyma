package configwatcher

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRuntimesService_GetRuntimes(t *testing.T) {
	fixRuntime1 := fixRuntime("runtime1", baseNamespace, "foo")
	fixRuntime2 := fixRuntime("runtime2", baseNamespace, "bar")
	fixRuntime3 := fixRuntime("runtime1", "foo", "bar")
	fixRuntime4 := fixRuntime("runtime2", "bar", "bar")

	t.Run("Success", func(t *testing.T) {
		service := fixRuntimesService(fixRuntime1, fixRuntime2, fixRuntime3, fixRuntime4)
		expected := map[string]*corev1.ConfigMap{
			"foo": fixRuntime1,
			"bar": fixRuntime2,
		}

		runtimes, err := service.GetRuntimes()
		require.NoError(t, err)
		assert.Exactly(t, expected, runtimes)
	})

	t.Run("Not found", func(t *testing.T) {
		service := fixRuntimesService(fixRuntime3, fixRuntime4)
		_, err := service.GetRuntimes()
		require.Error(t, err)
	})
}

func TestRuntimesService_UpdateCachedRuntime(t *testing.T) {
	fixRuntime1 := fixRuntime("runtime1", baseNamespace, "foo")
	fixRuntime2 := fixRuntime("runtime1", baseNamespace, "foo")
	fixRuntime2.Data = map[string]string{
		"foo": "foo",
		"bar": "bar",
	}

	t.Run("Success", func(t *testing.T) {
		service := fixRuntimesService(fixRuntime1)
		err := service.UpdateCachedRuntime(fixRuntime2)
		require.NoError(t, err)

		runtimes, err := service.GetRuntimes()
		require.NoError(t, err)
		assert.Exactly(t, fixRuntime2, runtimes["foo"])
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixRuntimesService(fixRuntime1)
		err := service.UpdateCachedRuntime(nil)
		require.Error(t, err)
	})

	t.Run("Not found", func(t *testing.T) {
		service := fixRuntimesService()
		err := service.UpdateCachedRuntime(fixRuntime2)
		require.Error(t, err)
	})
}

func TestCredentialsService_HandleRuntimeInNamespace(t *testing.T) {
	fixRuntime1 := fixRuntime("runtime1", baseNamespace, "foo")
	fixRuntime2 := fixRuntime("runtime1", "foo", "foo")
	fixRuntime2.Data = map[string]string{
		"foo": "foo",
		"bar": "bar",
	}

	t.Run("Success", func(t *testing.T) {
		service := fixRuntimesService(fixRuntime1, fixRuntime2)
		err := service.HandleRuntimesInNamespace("foo")
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, RuntimeLabelValue)
		list, err := service.coreClient.ConfigMaps("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixRuntime1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixRuntime1)
	})

	t.Run("Error", func(t *testing.T) {
		service := fixRuntimesService()
		err := service.HandleRuntimesInNamespace("foo")
		require.Error(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		fixRuntime1.Namespace = baseNamespace
		service := fixRuntimesService(fixRuntime1)
		err := service.HandleRuntimesInNamespace("foo")
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, RuntimeLabelValue)
		list, err := service.coreClient.ConfigMaps("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixRuntime1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixRuntime1)
	})
}

func TestCredentialsService_HandleRuntimeInNamespaces(t *testing.T) {
	fixRuntime1 := fixRuntime("runtime1", baseNamespace, "foo")
	fixRuntime2 := fixRuntime("runtime1", "foo", "foo")
	fixRuntime2.Data = map[string]string{
		"foo": "foo",
		"bar": "bar",
	}
	fixRuntime3 := fixRuntime("runtime1", "bar", "foo")
	fixRuntime3.Data = map[string]string{
		"foo": "foo",
		"bar": "bar",
	}

	t.Run("Success - if exist", func(t *testing.T) {
		service := fixRuntimesService(fixRuntime1, fixRuntime2, fixRuntime3)
		err := service.HandleRuntimeInNamespaces(fixRuntime1, []string{"foo", "bar"})
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, RuntimeLabelValue)
		list, err := service.coreClient.ConfigMaps("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixRuntime1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixRuntime1)

		labelSelector = fmt.Sprintf("%s=%s", ConfigLabel, RuntimeLabelValue)
		list, err = service.coreClient.ConfigMaps("bar").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixRuntime1.Namespace = "bar"
		assert.Exactly(t, &list.Items[0], fixRuntime1)
	})

	t.Run("Success - if not exist", func(t *testing.T) {
		service := fixRuntimesService(fixRuntime1)
		err := service.HandleRuntimeInNamespaces(fixRuntime1, []string{"foo", "bar"})
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, RuntimeLabelValue)
		list, err := service.coreClient.ConfigMaps("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixRuntime1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixRuntime1)

		labelSelector = fmt.Sprintf("%s=%s", ConfigLabel, RuntimeLabelValue)
		list, err = service.coreClient.ConfigMaps("bar").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixRuntime1.Namespace = "bar"
		assert.Exactly(t, &list.Items[0], fixRuntime1)
	})
}

func TestRuntimesService_IsBaseRuntime(t *testing.T) {
	fixRuntime1 := fixRuntime("runtime1", baseNamespace, "foo")
	fixRuntime2 := fixRuntime("runtime1", "foo", "bar")

	t.Run("True", func(t *testing.T) {
		service := fixRuntimesService()
		is := service.IsBaseRuntime(fixRuntime1)
		require.True(t, is)
	})

	t.Run("False", func(t *testing.T) {
		service := fixRuntimesService()
		is := service.IsBaseRuntime(fixRuntime2)
		require.False(t, is)
	})
}

func fixRuntimesService(objects ...runtime.Object) *RuntimesService {
	client := fixFakeClientset(objects...)
	return NewRuntimesService(client.CoreV1(), Config{
		BaseNamespace: baseNamespace,
	})
}

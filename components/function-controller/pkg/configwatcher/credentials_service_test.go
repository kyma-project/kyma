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

func TestCredentialsService_GetCredentials(t *testing.T) {
	fixCredential1 := fixCredential("credential1", baseNamespace, "foo")
	fixCredential2 := fixCredential("credential2", baseNamespace, "bar")
	fixCredential3 := fixCredential("credential1", "foo", "bar")
	fixCredential4 := fixCredential("credential2", "bar", "bar")

	t.Run("Success", func(t *testing.T) {
		service := fixCredentialsService(fixCredential1, fixCredential2, fixCredential3, fixCredential4)
		expected := map[string]*corev1.Secret{
			"foo": fixCredential1,
			"bar": fixCredential2,
		}

		credentials, err := service.GetCredentials()
		require.NoError(t, err)
		assert.Exactly(t, expected, credentials)
	})

	t.Run("Not found", func(t *testing.T) {
		service := fixCredentialsService(fixCredential3, fixCredential4)
		_, err := service.GetCredentials()
		require.Error(t, err)
	})
}

func TestCredentialsService_GetCredential(t *testing.T) {
	fixCredential1 := fixCredential("credential1", baseNamespace, "foo")
	fixCredential2 := fixCredential("credential2", baseNamespace, "bar")
	fixCredential3 := fixCredential("credential1", "foo", "bar")
	fixCredential4 := fixCredential("credential2", "bar", "bar")

	t.Run("Success", func(t *testing.T) {
		service := fixCredentialsService(fixCredential1, fixCredential2, fixCredential3, fixCredential4)
		credential, err := service.GetCredential("foo")
		require.NoError(t, err)
		assert.Exactly(t, fixCredential1, credential)
	})

	t.Run("Not found", func(t *testing.T) {
		service := fixCredentialsService(fixCredential3, fixCredential4)
		_, err := service.GetCredential("foo")
		require.Error(t, err)
	})
}

func TestCredentialsService_UpdateCachedCredential(t *testing.T) {
	fixCredential1 := fixCredential("credential1", baseNamespace, "foo")
	fixCredential2 := fixCredential("credential1", baseNamespace, "foo")
	fixCredential2.Data = map[string][]byte{
		"foo": []byte(`foo`),
		"bar": []byte(`bar`),
	}

	t.Run("Success", func(t *testing.T) {
		service := fixCredentialsService(fixCredential1)
		err := service.UpdateCachedCredential(fixCredential2)
		require.NoError(t, err)

		credential, err := service.GetCredential("foo")
		require.NoError(t, err)
		assert.Exactly(t, fixCredential2, credential)
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixCredentialsService(fixCredential1)
		err := service.UpdateCachedCredential(nil)
		require.Error(t, err)
	})

	t.Run("Not found", func(t *testing.T) {
		service := fixCredentialsService()
		err := service.UpdateCachedCredential(fixCredential2)
		require.Error(t, err)
	})
}

func TestCredentialsService_CreateCredentialsInNamespace(t *testing.T) {
	fixCredential1 := fixCredential("credential1", baseNamespace, "foo")
	fixCredential2 := fixCredential("credential1", "foo", "foo")

	t.Run("Success", func(t *testing.T) {
		service := fixCredentialsService(fixCredential1)
		err := service.CreateCredentialsInNamespace("foo")
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, CredentialsLabelValue)
		list, err := service.coreClient.Secrets("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		assert.Exactly(t, &list.Items[0], fixCredential2)
	})

	t.Run("Error", func(t *testing.T) {
		service := fixCredentialsService()
		err := service.CreateCredentialsInNamespace("foo")
		require.Error(t, err)
	})

	t.Run("Already exists", func(t *testing.T) {
		service := fixCredentialsService(fixCredential1)
		err := service.CreateCredentialsInNamespace("foo")
		require.NoError(t, err)
		err = service.CreateCredentialsInNamespace("foo")
		require.NoError(t, err)
	})
}

func TestCredentialsService_UpdateCredentialsInNamespace(t *testing.T) {
	fixCredential1 := fixCredential("credential1", baseNamespace, "foo")
	fixCredential2 := fixCredential("credential1", "foo", "foo")
	fixCredential2.Data = map[string][]byte{
		"foo": []byte(`foo`),
		"bar": []byte(`bar`),
	}

	t.Run("Success", func(t *testing.T) {
		service := fixCredentialsService(fixCredential1, fixCredential2)
		err := service.UpdateCredentialsInNamespace("foo")
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, CredentialsLabelValue)
		list, err := service.coreClient.Secrets("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixCredential1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixCredential1)
	})

	t.Run("Error", func(t *testing.T) {
		service := fixCredentialsService()
		err := service.UpdateCredentialsInNamespace("foo")
		require.Error(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		fixCredential1.Namespace = baseNamespace
		service := fixCredentialsService(fixCredential1)
		err := service.UpdateCredentialsInNamespace("foo")
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, CredentialsLabelValue)
		list, err := service.coreClient.Secrets("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixCredential1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixCredential1)
	})
}

func TestCredentialsService_UpdateCredentialInNamespaces(t *testing.T) {
	fixCredential1 := fixCredential("credential1", baseNamespace, "foo")
	fixCredential2 := fixCredential("credential1", "foo", "foo")
	fixCredential2.Data = map[string][]byte{
		"foo": []byte(`foo`),
		"bar": []byte(`bar`),
	}
	fixCredential3 := fixCredential("credential1", "bar", "foo")
	fixCredential3.Data = map[string][]byte{
		"foo": []byte(`foo`),
		"bar": []byte(`bar`),
	}

	t.Run("Success - if exist", func(t *testing.T) {
		service := fixCredentialsService(fixCredential1, fixCredential2, fixCredential3)
		err := service.UpdateCredentialInNamespaces(fixCredential1, []string{"foo", "bar"})
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, CredentialsLabelValue)
		list, err := service.coreClient.Secrets("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixCredential1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixCredential1)

		labelSelector = fmt.Sprintf("%s=%s", ConfigLabel, CredentialsLabelValue)
		list, err = service.coreClient.Secrets("bar").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixCredential1.Namespace = "bar"
		assert.Exactly(t, &list.Items[0], fixCredential1)
	})

	t.Run("Success - if not exists", func(t *testing.T) {
		service := fixCredentialsService(fixCredential1)
		err := service.UpdateCredentialInNamespaces(fixCredential1, []string{"foo", "bar"})
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, CredentialsLabelValue)
		list, err := service.coreClient.Secrets("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixCredential1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixCredential1)

		labelSelector = fmt.Sprintf("%s=%s", ConfigLabel, CredentialsLabelValue)
		list, err = service.coreClient.Secrets("bar").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixCredential1.Namespace = "bar"
		assert.Exactly(t, &list.Items[0], fixCredential1)
	})
}

func TestCredentialsService_IsBaseCredential(t *testing.T) {
	fixCredential1 := fixCredential("credential1", baseNamespace, "foo")
	fixCredential2 := fixCredential("credential1", "foo", "bar")

	t.Run("True", func(t *testing.T) {
		service := fixCredentialsService()
		is := service.IsBaseCredential(fixCredential1)
		require.True(t, is)
	})

	t.Run("False", func(t *testing.T) {
		service := fixCredentialsService()
		is := service.IsBaseCredential(fixCredential2)
		require.False(t, is)
	})
}

func fixCredentialsService(objects ...runtime.Object) *CredentialsService {
	client := fixFakeClientset(objects...)
	return NewCredentialsService(client.CoreV1(), Config{
		BaseNamespace: baseNamespace,
	})
}

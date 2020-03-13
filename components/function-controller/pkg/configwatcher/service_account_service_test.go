package configwatcher

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestServiceAccountService_GetServiceAccount(t *testing.T) {
	fixSA1 := fixServiceAccount("sa1", baseNamespace, "credential1", "credential2")
	fixSA2 := fixServiceAccount("sa2", "foo", "credential1", "credential2")

	t.Run("Success", func(t *testing.T) {
		service := fixServiceAccountService(fixSA1, fixSA2)
		sa, err := service.GetServiceAccount()
		require.NoError(t, err)
		assert.Exactly(t, fixSA1, sa)
	})

	t.Run("Not found", func(t *testing.T) {
		service := fixServiceAccountService(fixSA2)
		_, err := service.GetServiceAccount()
		require.Error(t, err)
	})
}

func TestServiceAccountService_HandleServiceAccountInNamespace(t *testing.T) {
	fixSA1 := fixServiceAccount("sa1", baseNamespace, "credential1", "credential2")
	fixSA2 := fixServiceAccount("sa1", "foo", "credential1", "credential3")

	t.Run("Success", func(t *testing.T) {
		service := fixServiceAccountService(fixSA1, fixSA2)
		err := service.HandleServiceAccountInNamespace("foo")
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, ServiceAccountLabelValue)
		list, err := service.coreClient.ServiceAccounts("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixSA1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixSA1)
	})

	t.Run("Error", func(t *testing.T) {
		service := fixServiceAccountService()
		err := service.HandleServiceAccountInNamespace("foo")
		require.Error(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		fixSA1.Namespace = baseNamespace
		service := fixServiceAccountService(fixSA1)
		err := service.HandleServiceAccountInNamespace("foo")
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, ServiceAccountLabelValue)
		list, err := service.coreClient.ServiceAccounts("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixSA1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixSA1)
	})
}

func TestServiceAccountService_HandleServiceAccountInNamespaces(t *testing.T) {
	fixSA1 := fixServiceAccount("sa1", baseNamespace, "credential1", "credential2")
	fixSA2 := fixServiceAccount("sa1", "foo", "credential1", "credential3")
	fixSA3 := fixServiceAccount("sa1", "bar", "credential3", "credential2")

	t.Run("Success - if exist", func(t *testing.T) {
		service := fixServiceAccountService(fixSA1, fixSA2, fixSA3)
		err := service.HandleServiceAccountInNamespaces(fixSA1, []string{"foo", "bar"})
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, ServiceAccountLabelValue)
		list, err := service.coreClient.ServiceAccounts("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixSA1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixSA1)

		labelSelector = fmt.Sprintf("%s=%s", ConfigLabel, ServiceAccountLabelValue)
		list, err = service.coreClient.ServiceAccounts("bar").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixSA1.Namespace = "bar"
		assert.Exactly(t, &list.Items[0], fixSA1)
	})

	t.Run("Success - if not exist", func(t *testing.T) {
		service := fixServiceAccountService(fixSA1)
		err := service.HandleServiceAccountInNamespaces(fixSA1, []string{"foo", "bar"})
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, ServiceAccountLabelValue)
		list, err := service.coreClient.ServiceAccounts("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixSA1.Namespace = "foo"
		assert.Exactly(t, &list.Items[0], fixSA1)

		labelSelector = fmt.Sprintf("%s=%s", ConfigLabel, ServiceAccountLabelValue)
		list, err = service.coreClient.ServiceAccounts("bar").List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		fixSA1.Namespace = "bar"
		assert.Exactly(t, &list.Items[0], fixSA1)
	})
}

func TestServiceAccountService_IsBaseServiceAccount(t *testing.T) {
	fixSA1 := fixServiceAccount("sa1", baseNamespace, "credential1", "credential2")
	fixSA2 := fixServiceAccount("sa2", "foo", "credential1", "credential2")

	t.Run("True", func(t *testing.T) {
		service := fixServiceAccountService()
		is := service.IsBaseServiceAccount(fixSA1)
		require.True(t, is)
	})

	t.Run("False", func(t *testing.T) {
		service := fixServiceAccountService()
		is := service.IsBaseServiceAccount(fixSA2)
		require.False(t, is)
	})
}

func fixServiceAccountService(objects ...runtime.Object) *ServiceAccountService {
	client := fixFakeClientset(objects...)
	config := Config{
		BaseNamespace: baseNamespace,
	}
	return NewServiceAccountService(client.CoreV1(), config)
}

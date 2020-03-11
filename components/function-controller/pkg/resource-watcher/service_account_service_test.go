package resource_watcher

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

func TestServiceAccountService_CreateServiceAccountInNamespace(t *testing.T) {
	fixCredential1 := fixCredential("credential1", baseNamespace, RegistryCredentialsLabelValue)
	fixCredential2 := fixCredential("credential2", baseNamespace, ImagePullSecretLabelValue)
	fixSA1 := fixServiceAccount("sa1", baseNamespace, "credential1", "credential2")
	fixSA2 := fixServiceAccount("sa1", "foo", "credential1", "credential2")

	t.Run("Success", func(t *testing.T) {
		service := fixServiceAccountService(fixSA1, fixCredential1, fixCredential2)
		err := service.CreateServiceAccountInNamespace("foo")
		require.NoError(t, err)

		labelSelector := fmt.Sprintf("%s=%s", ConfigLabel, ServiceAccountLabelValue)
		list, err := service.coreClient.ServiceAccounts("foo").List(metav1.ListOptions{
			LabelSelector: labelSelector,
			Limit:         1,
		})
		require.NoError(t, err)
		assert.Len(t, list.Items, 1)
		assert.Exactly(t, &list.Items[0], fixSA2)
	})

	t.Run("Error", func(t *testing.T) {
		service := fixServiceAccountService()
		err := service.CreateServiceAccountInNamespace("foo")
		require.Error(t, err)
	})

	t.Run("Already exists", func(t *testing.T) {
		service := fixServiceAccountService(fixSA1, fixCredential1, fixCredential2)
		err := service.CreateServiceAccountInNamespace("foo")
		require.NoError(t, err)
		err = service.CreateServiceAccountInNamespace("foo")
		require.NoError(t, err)
	})
}

func fixServiceAccountService(objects ...runtime.Object) *ServiceAccountService {
	client := fixFakeClientset(objects...)
	config := Config{
		BaseNamespace: baseNamespace,
	}
	credentialsService := NewCredentialsService(client.CoreV1(), config)
	return NewServiceAccountService(client.CoreV1(), config, credentialsService)
}

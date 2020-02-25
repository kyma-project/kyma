package k8s_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	appAutomock "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespaceResolver_NamespacesQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name, inactiveName, systemName := "name", "inactive", "system"

		k8sNamespace := fixNamespaceWithStatus(name, "Active")
		gqlNamespace := gqlschema.Namespace{
			Name:              name,
			Status:            "Active",
			IsSystemNamespace: false,
		}
		k8sInactiveNamespace := fixNamespaceWithStatus(inactiveName, "Terminating")
		gqlInactiveNamespace := gqlschema.Namespace{
			Name:              inactiveName,
			Status:            "Terminating",
			IsSystemNamespace: false,
		}
		k8sSystemNamespace := fixNamespaceWithStatus(systemName, "Active")
		gqlSystemNamespace := gqlschema.Namespace{
			Name:              systemName,
			Status:            "Active",
			IsSystemNamespace: true,
		}

		resources := []*v1.Namespace{k8sNamespace, k8sInactiveNamespace, k8sSystemNamespace}

		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("List").Return(resources, nil).Times(3)
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{systemName})

		// check with default values
		result, err := resolver.NamespacesQuery(nil, nil, nil)
		expected := []gqlschema.Namespace{
			gqlNamespace,
		}

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		trueBool := true

		// check with system namespaces
		result, err = resolver.NamespacesQuery(nil, &trueBool, nil)
		expected = []gqlschema.Namespace{
			gqlNamespace, gqlSystemNamespace,
		}

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		// check with inactive namespaces
		result, err = resolver.NamespacesQuery(nil, nil, &trueBool)
		expected = []gqlschema.Namespace{
			gqlNamespace, gqlInactiveNamespace,
		}

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		resources := []*v1.Namespace{}
		expected := []gqlschema.Namespace(nil)
		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("List").Return(resources, nil).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		result, err := resolver.NamespacesQuery(nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorListing", func(t *testing.T) {
		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("List").Return(nil, errors.New("test error")).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		result, err := resolver.NamespacesQuery(nil, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestNamespaceResolver_NamespaceQuery(t *testing.T) {
	name := "name"
	labels := map[string]string{
		"env": "true",
	}

	t.Run("Success", func(t *testing.T) {
		resource := fixNamespace(name, labels)
		expected := gqlschema.Namespace{
			Name:   name,
			Labels: labels,
		}

		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Find", name).Return(resource, nil).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		result, err := resolver.NamespaceQuery(nil, name)

		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Find", name).Return(nil, nil).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		result, err := resolver.NamespaceQuery(nil, name)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error finding", func(t *testing.T) {
		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Find", name).Return(nil, errors.New("test error")).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		result, err := resolver.NamespaceQuery(nil, name)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestNamespaceResolver_CreateNamespace(t *testing.T) {
	name := "exampleName"

	t.Run("Success", func(t *testing.T) {
		labels := gqlschema.Labels{
			"test": "true",
		}

		resource := fixNamespace(name, labels)
		expected := gqlschema.NamespaceMutationOutput{
			Name:   name,
			Labels: labels,
		}

		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Create", name, labels).Return(resource, nil).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		result, err := resolver.CreateNamespace(nil, name, &labels)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		labels := gqlschema.Labels{}

		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Create", name, labels).Return(nil, errors.New("test error")).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		result, err := resolver.CreateNamespace(nil, name, &labels)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.NotNil(t, result)
	})
}

func TestNamespaceResolver_UpdateNamespace(t *testing.T) {
	name := "exampleName"
	labels := gqlschema.Labels{
		"test": "true",
	}

	t.Run("Success", func(t *testing.T) {
		resource := fixNamespace(name, labels)
		expected := gqlschema.NamespaceMutationOutput{
			Name:   name,
			Labels: labels,
		}

		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Update", name, labels).Return(resource, nil).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		result, err := resolver.UpdateNamespace(nil, name, labels)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Update", name, labels).Return(nil, errors.New("test error")).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		result, err := resolver.UpdateNamespace(nil, name, labels)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.NotNil(t, result)
	})
}

func TestNamespaceResolver_DeleteNamespace(t *testing.T) {
	name := "name"
	labels := map[string]string{
		"env": "true",
	}
	expected := gqlschema.Namespace{
		Name:   name,
		Labels: labels,
	}

	t.Run("Success", func(t *testing.T) {
		resource := fixNamespace(name, labels)

		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Find", name).Return(resource, nil).Once()
		svc.On("Delete", name).Return(nil).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		result, err := resolver.DeleteNamespace(nil, name)

		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Error finding", func(t *testing.T) {
		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Find", name).Return(nil, errors.New("test error")).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		result, err := resolver.DeleteNamespace(nil, name)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("Error deleting", func(t *testing.T) {
		resource := fixNamespace(name, labels)

		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Find", name).Return(resource, nil).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		svc.On("Delete", name).Return(errors.New("test error")).Once()

		result, err := resolver.DeleteNamespace(nil, name)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestNamespaceResolver_NamespaceEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), -24*time.Hour)
		cancel()

		svc := automock.NewNamespaceSvc()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()

		appRetriever := new(appAutomock.ApplicationRetriever)
		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		_, err := resolver.NamespaceEventSubscription(ctx, nil)
		require.NoError(t, err)

		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), -24*time.Hour)
		cancel()

		svc := automock.NewNamespaceSvc()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()

		appRetriever := new(appAutomock.ApplicationRetriever)
		resolver := k8s.NewNamespaceResolver(svc, appRetriever, []string{})

		channel, err := resolver.NamespaceEventSubscription(ctx, nil)
		require.NoError(t, err)

		_, ok := <-channel
		assert.False(t, ok)

		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}

func fixNamespaceWithStatus(name string, status string) *v1.Namespace {
	namespace := fixNamespaceWithoutTypeMeta(name, nil)
	namespace.TypeMeta = metav1.TypeMeta{
		Kind:       "Namespace",
		APIVersion: "v1",
	}

	namespace.Status.Phase = v1.NamespacePhase(status)
	return namespace
}

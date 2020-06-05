package k8s_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sTesting "k8s.io/client-go/testing"
)

func failingReactor(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
	return true, nil, errors.New("custom error")
}

func TestSecretResolver_SecretQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t1 := time.Unix(1, 0)
		expected := &gqlschema.Secret{
			Name:         "Test",
			Namespace:    "TestNS",
			CreationTime: t1,
			Annotations:  gqlschema.JSON{"second-annot": "content"},
		}

		name := "name"
		namespace := "namespace"
		resource := &v1.Secret{}
		resourceGetter := automock.NewSecretSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLSecretConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(resourceGetter)
		resolver.SetSecretConverter(converter)

		result, err := resolver.SecretQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		namespace := "namespace"
		resourceGetter := automock.NewSecretSvc()
		resourceGetter.On("Find", name, namespace).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(resourceGetter)

		result, err := resolver.SecretQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Nil(t, result)
	})
	t.Run("ErrorGetting", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"
		namespace := "namespace"
		resource := &v1.Secret{}
		resourceGetter := automock.NewSecretSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(resourceGetter)

		result, err := resolver.SecretQuery(nil, name, namespace)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestSecretResolver_SecretsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "Test"
		namespace := "namespace"
		resource := &v1.Secret{}
		resources := []*v1.Secret{
			resource, resource,
		}
		expected := []gqlschema.Secret{
			{
				Name: name,
			},
			{
				Name: name,
			},
		}

		resourceGetter := automock.NewSecretSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLSecretConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(resourceGetter)
		resolver.SetSecretConverter(converter)

		result, err := resolver.SecretsQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		namespace := "namespace"
		var resources []*v1.Secret
		var expected []gqlschema.Secret

		resourceGetter := automock.NewSecretSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(resourceGetter)

		result, err := resolver.SecretsQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		namespace := "namespace"
		expected := errors.New("Test")
		var resources []*v1.Secret
		resourceGetter := automock.NewSecretSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(resourceGetter)

		result, err := resolver.SecretsQuery(nil, namespace, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestSecretResolver_SecretEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewSecretSvc()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := k8s.NewSecretResolver(svc)

		_, err := resolver.SecretEventSubscription(ctx, "test")

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewSecretSvc()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := k8s.NewSecretResolver(svc)

		channel, err := resolver.SecretEventSubscription(ctx, "test")
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}

func TestSecretResolver_UpdateSecretMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedSecretFix := fixSecret(name, namespace, map[string]string{
			"test": "test",
		})
		updatedGQLSecretFix := &gqlschema.Secret{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]interface{}{
				"test": "test",
			},
		}
		gqlJSONFix := gqlschema.JSON{}

		secretSvc := automock.NewSecretSvc()
		secretSvc.On("Update", name, namespace, *updatedSecretFix).Return(updatedSecretFix, nil).Once()
		defer secretSvc.AssertExpectations(t)

		converter := automock.NewGQLSecretConverter()
		converter.On("GQLJSONToSecret", gqlJSONFix).Return(*updatedSecretFix, nil).Once()
		converter.On("ToGQL", updatedSecretFix).Return(updatedGQLSecretFix, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(secretSvc)
		resolver.SetSecretConverter(converter)

		result, err := resolver.UpdateSecretMutation(nil, name, namespace, gqlJSONFix)

		require.NoError(t, err)
		assert.Equal(t, updatedGQLSecretFix, result)
	})

	t.Run("ErrorConvertingToSecret", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		secretSvc := automock.NewSecretSvc()

		converter := automock.NewGQLSecretConverter()
		converter.On("GQLJSONToSecret", gqlJSONFix).Return(v1.Secret{}, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(secretSvc)
		resolver.SetSecretConverter(converter)

		result, err := resolver.UpdateSecretMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorUpdating", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedSecretFix := fixSecret(name, namespace, map[string]string{
			"test": "test",
		})
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		secretSvc := automock.NewSecretSvc()
		secretSvc.On("Update", name, namespace, *updatedSecretFix).Return(nil, expected).Once()
		defer secretSvc.AssertExpectations(t)

		converter := automock.NewGQLSecretConverter()
		converter.On("GQLJSONToSecret", gqlJSONFix).Return(*updatedSecretFix, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(secretSvc)
		resolver.SetSecretConverter(converter)

		result, err := resolver.UpdateSecretMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSecretResolver_DeleteSecretMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixSecret(name, namespace, nil)
		expected := &gqlschema.Secret{
			Name:      name,
			Namespace: namespace,
		}

		secretSvc := automock.NewSecretSvc()
		secretSvc.On("Find", name, namespace).Return(resource, nil).Once()
		secretSvc.On("Delete", name, namespace).Return(nil).Once()
		defer secretSvc.AssertExpectations(t)

		converter := automock.NewGQLSecretConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(secretSvc)
		resolver.SetSecretConverter(converter)

		result, err := resolver.DeleteSecretMutation(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		expected := errors.New("fix")

		secretSvc := automock.NewSecretSvc()
		secretSvc.On("Find", name, namespace).Return(nil, expected).Once()
		defer secretSvc.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(secretSvc)

		result, err := resolver.DeleteSecretMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorDeleting", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixSecret(name, namespace, nil)
		expected := errors.New("fix")

		secretSvc := automock.NewSecretSvc()
		secretSvc.On("Find", name, namespace).Return(resource, nil).Once()
		secretSvc.On("Delete", name, namespace).Return(expected).Once()
		defer secretSvc.AssertExpectations(t)

		resolver := k8s.NewSecretResolver(secretSvc)

		result, err := resolver.DeleteSecretMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

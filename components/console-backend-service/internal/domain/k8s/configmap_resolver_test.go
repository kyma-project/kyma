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
)

func TestConfigMapResolver_ConfigMapQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.ConfigMap{
			Name: "Test",
		}
		name := "name"
		namespace := "namespace"
		resource := &v1.ConfigMap{}
		resourceGetter := automock.NewConfigMapSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlConfigMapConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(resourceGetter)
		resolver.SetConfigMapConverter(converter)

		result, err := resolver.ConfigMapQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		namespace := "namespace"
		resourceGetter := automock.NewConfigMapSvc()
		resourceGetter.On("Find", name, namespace).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(resourceGetter)

		result, err := resolver.ConfigMapQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"
		namespace := "namespace"
		resource := &v1.ConfigMap{}
		resourceGetter := automock.NewConfigMapSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(resourceGetter)

		result, err := resolver.ConfigMapQuery(nil, name, namespace)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConverting", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"
		namespace := "namespace"
		resource := &v1.ConfigMap{}
		resourceGetter := automock.NewConfigMapSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlConfigMapConverter()
		converter.On("ToGQL", resource).Return(nil, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(resourceGetter)
		resolver.SetConfigMapConverter(converter)

		result, err := resolver.ConfigMapQuery(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestConfigMapResolver_ConfigMapsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "Test"
		namespace := "namespace"
		resource := fixConfigMap(name, namespace, map[string]string{
			"test": "test",
		})
		resources := []*v1.ConfigMap{
			resource, resource,
		}
		expected := []gqlschema.ConfigMap{
			{
				Name: name,
			},
			{
				Name: name,
			},
		}

		resourceGetter := automock.NewConfigMapSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlConfigMapConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(resourceGetter)
		resolver.SetConfigMapConverter(converter)

		result, err := resolver.ConfigMapsQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		namespace := "namespace"
		var resources []*v1.ConfigMap
		var expected []gqlschema.ConfigMap

		resourceGetter := automock.NewConfigMapSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(resourceGetter)

		result, err := resolver.ConfigMapsQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		namespace := "namespace"
		expected := errors.New("Test")
		var resources []*v1.ConfigMap
		resourceGetter := automock.NewConfigMapSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(resourceGetter)

		result, err := resolver.ConfigMapsQuery(nil, namespace, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConverting", func(t *testing.T) {
		name := "Test"
		namespace := "namespace"
		resource := fixConfigMap(name, namespace, map[string]string{
			"test": "test",
		})
		resources := []*v1.ConfigMap{
			resource, resource,
		}
		expected := errors.New("Test")

		resourceGetter := automock.NewConfigMapSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlConfigMapConverter()
		converter.On("ToGQLs", resources).Return(nil, expected)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(resourceGetter)
		resolver.SetConfigMapConverter(converter)

		result, err := resolver.ConfigMapsQuery(nil, namespace, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestConfigMapResolver_UpdateConfigMapMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedConfigMapFix := fixConfigMap(name, namespace, map[string]string{
			"test": "test",
		})
		updatedGQLConfigMapFix := &gqlschema.ConfigMap{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"test": "test",
			},
		}
		gqlJSONFix := gqlschema.JSON{}

		configMapSvc := automock.NewConfigMapSvc()
		configMapSvc.On("Update", name, namespace, *updatedConfigMapFix).Return(updatedConfigMapFix, nil).Once()
		defer configMapSvc.AssertExpectations(t)

		converter := automock.NewGqlConfigMapConverter()
		converter.On("GQLJSONToConfigMap", gqlJSONFix).Return(*updatedConfigMapFix, nil).Once()
		converter.On("ToGQL", updatedConfigMapFix).Return(updatedGQLConfigMapFix, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(configMapSvc)
		resolver.SetConfigMapConverter(converter)

		result, err := resolver.UpdateConfigMapMutation(nil, name, namespace, gqlJSONFix)

		require.NoError(t, err)
		assert.Equal(t, updatedGQLConfigMapFix, result)
	})

	t.Run("ErrorConvertingToConfigMap", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		configMapSvc := automock.NewConfigMapSvc()

		converter := automock.NewGqlConfigMapConverter()
		converter.On("GQLJSONToConfigMap", gqlJSONFix).Return(v1.ConfigMap{}, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(configMapSvc)
		resolver.SetConfigMapConverter(converter)

		result, err := resolver.UpdateConfigMapMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorUpdating", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedConfigMapFix := fixConfigMap(name, namespace, map[string]string{
			"test": "test",
		})
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		configMapSvc := automock.NewConfigMapSvc()
		configMapSvc.On("Update", name, namespace, *updatedConfigMapFix).Return(nil, expected).Once()
		defer configMapSvc.AssertExpectations(t)

		converter := automock.NewGqlConfigMapConverter()
		converter.On("GQLJSONToConfigMap", gqlJSONFix).Return(*updatedConfigMapFix, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(configMapSvc)
		resolver.SetConfigMapConverter(converter)

		result, err := resolver.UpdateConfigMapMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ErrorConvertingToGQL", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedConfigMapFix := fixConfigMap(name, namespace, map[string]string{
			"test": "test",
		})
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		configMapSvc := automock.NewConfigMapSvc()
		configMapSvc.On("Update", name, namespace, *updatedConfigMapFix).Return(updatedConfigMapFix, nil).Once()
		defer configMapSvc.AssertExpectations(t)

		converter := automock.NewGqlConfigMapConverter()
		converter.On("GQLJSONToConfigMap", gqlJSONFix).Return(*updatedConfigMapFix, nil).Once()
		converter.On("ToGQL", updatedConfigMapFix).Return(nil, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(configMapSvc)
		resolver.SetConfigMapConverter(converter)

		result, err := resolver.UpdateConfigMapMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestConfigMapResolver_DeleteConfigMapMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixConfigMap(name, namespace, map[string]string{
			"test": "test",
		})
		expected := &gqlschema.ConfigMap{
			Name:      name,
			Namespace: namespace,
		}

		configMapSvc := automock.NewConfigMapSvc()
		configMapSvc.On("Find", name, namespace).Return(resource, nil).Once()
		configMapSvc.On("Delete", name, namespace).Return(nil).Once()
		defer configMapSvc.AssertExpectations(t)

		converter := automock.NewGqlConfigMapConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(configMapSvc)
		resolver.SetConfigMapConverter(converter)

		result, err := resolver.DeleteConfigMapMutation(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		expected := errors.New("fix")

		configMapSvc := automock.NewConfigMapSvc()
		configMapSvc.On("Find", name, namespace).Return(nil, expected).Once()
		defer configMapSvc.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(configMapSvc)

		result, err := resolver.DeleteConfigMapMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorDeleting", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixConfigMap(name, namespace, map[string]string{
			"test": "test",
		})
		expected := errors.New("fix")

		configMapSvc := automock.NewConfigMapSvc()
		configMapSvc.On("Find", name, namespace).Return(resource, nil).Once()
		configMapSvc.On("Delete", name, namespace).Return(expected).Once()
		defer configMapSvc.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(configMapSvc)

		result, err := resolver.DeleteConfigMapMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConverting", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixConfigMap(name, namespace, map[string]string{
			"test": "test",
		})
		error := errors.New("fix")

		configMapSvc := automock.NewConfigMapSvc()
		configMapSvc.On("Find", name, namespace).Return(resource, nil).Once()
		defer configMapSvc.AssertExpectations(t)

		converter := automock.NewGqlConfigMapConverter()
		converter.On("ToGQL", resource).Return(nil, error).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewConfigMapResolver(configMapSvc)
		resolver.SetConfigMapConverter(converter)

		result, err := resolver.DeleteConfigMapMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestConfigMapResolver_ConfigMapEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewConfigMapSvc()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := k8s.NewConfigMapResolver(svc)

		_, err := resolver.ConfigMapEventSubscription(ctx, "test")

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewConfigMapSvc()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := k8s.NewConfigMapResolver(svc)

		channel, err := resolver.ConfigMapEventSubscription(ctx, "test")
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}

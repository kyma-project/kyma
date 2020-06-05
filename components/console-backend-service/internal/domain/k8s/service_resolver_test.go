package k8s_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"

	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceResolver_ServiceQuery(t *testing.T) {
	name := "name"
	namespace := "namespace"

	assert := assert.New(t)

	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.Service{
			Name: "Test",
		}
		resource := &v1.Service{}
		resourceGetter := automock.NewServiceSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlServiceConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ServiceQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		resourceGetter := automock.NewServiceSvc()
		resourceGetter.On("Find", name, namespace).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(resourceGetter)

		result, err := resolver.ServiceQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Nil(result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		expected := errors.New("test")
		resource := &v1.Service{}
		resourceGetter := automock.NewServiceSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(resourceGetter)

		result, err := resolver.ServiceQuery(nil, name, namespace)

		assert.Error(err)
		assert.True(gqlerror.IsInternal(err))
		assert.Nil(result)
	})

}

func TestServiceResolver_ServicesQuery(t *testing.T) {
	name := "name"
	namespace := "namespace"

	assert := assert.New(t)

	t.Run("Success", func(t *testing.T) {
		resource := fixService(name, namespace, nil)
		resources := []*v1.Service{
			resource, resource,
		}
		expected := []gqlschema.Service{
			{
				Name: name,
			},
			{
				Name: name,
			},
		}

		resourceGetter := automock.NewServiceSvc()
		resourceGetter.On("List", namespace, []string(nil), pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlServiceConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ServicesQuery(nil, namespace, nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var resources []*v1.Service
		var expected []gqlschema.Service

		resourceGetter := automock.NewServiceSvc()
		resourceGetter.On("List", namespace, []string(nil), pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(resourceGetter)

		result, err := resolver.ServicesQuery(nil, namespace, nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(expected, result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		expected := errors.New("test")
		var resources []*v1.Service
		resourceGetter := automock.NewServiceSvc()
		resourceGetter.On("List", namespace, []string(nil), pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(resourceGetter)

		result, err := resolver.ServicesQuery(nil, namespace, nil, nil, nil)

		require.Error(t, err)
		assert.True(gqlerror.IsInternal(err))
		assert.Nil(result)
	})
}

func TestServiceResolver_ServiceEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewServiceSvc()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := k8s.NewServiceResolver(svc)

		_, err := resolver.ServiceEventSubscription(ctx, "test")

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewServiceSvc()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := k8s.NewServiceResolver(svc)

		channel, err := resolver.ServiceEventSubscription(ctx, "test")
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}

func TestServiceResolver_UpdateServiceMutation(t *testing.T) {
	assert := assert.New(t)
	t.Run("Success", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedServiceFix := fixService(name, namespace, map[string]string{
			"test": "test",
		})
		updatedGQLServiceFix := &gqlschema.Service{
			Name: name,
			Labels: map[string]string{
				"test": "test",
			},
		}
		gqlJSONFix := gqlschema.JSON{}

		serviceSvc := automock.NewServiceSvc()
		serviceSvc.On("Update", name, namespace, *updatedServiceFix).Return(updatedServiceFix, nil).Once()
		defer serviceSvc.AssertExpectations(t)

		converter := automock.NewGqlServiceConverter()
		converter.On("GQLJSONToService", gqlJSONFix).Return(*updatedServiceFix, nil).Once()
		converter.On("ToGQL", updatedServiceFix).Return(updatedGQLServiceFix, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(serviceSvc)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.UpdateServiceMutation(nil, name, namespace, gqlJSONFix)

		require.NoError(t, err)
		assert.Equal(updatedGQLServiceFix, result)
	})

	t.Run("ErrorConvertingToService", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		serviceSvc := automock.NewServiceSvc()

		converter := automock.NewGqlServiceConverter()
		converter.On("GQLJSONToService", gqlJSONFix).Return(v1.Service{}, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(serviceSvc)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.UpdateServiceMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.True(gqlerror.IsInternal(err))
		assert.Nil(result)
	})

	t.Run("ErrorUpdating", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedServiceFix := fixService(name, namespace, map[string]string{
			"test": "test",
		})
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		serviceSvc := automock.NewServiceSvc()
		serviceSvc.On("Update", name, namespace, *updatedServiceFix).Return(nil, expected).Once()
		defer serviceSvc.AssertExpectations(t)

		converter := automock.NewGqlServiceConverter()
		converter.On("GQLJSONToService", gqlJSONFix).Return(*updatedServiceFix, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(serviceSvc)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.UpdateServiceMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.Nil(result)
	})
}

func TestServiceResolver_DeleteServiceMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixService(name, namespace, nil)
		expected := &gqlschema.Service{
			Name: name,
		}

		serviceSvc := automock.NewServiceSvc()
		serviceSvc.On("Find", name, namespace).Return(resource, nil).Once()
		serviceSvc.On("Delete", name, namespace).Return(nil).Once()
		defer serviceSvc.AssertExpectations(t)

		converter := automock.NewGqlServiceConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(serviceSvc)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.DeleteServiceMutation(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		expected := errors.New("fix")

		serviceSvc := automock.NewServiceSvc()
		serviceSvc.On("Find", name, namespace).Return(nil, expected).Once()
		defer serviceSvc.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(serviceSvc)

		result, err := resolver.DeleteServiceMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorDeleting", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixService(name, namespace, nil)
		expected := errors.New("fix")

		serviceSvc := automock.NewServiceSvc()
		serviceSvc.On("Find", name, namespace).Return(resource, nil).Once()
		serviceSvc.On("Delete", name, namespace).Return(expected).Once()
		defer serviceSvc.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(serviceSvc)

		result, err := resolver.DeleteServiceMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConverting", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixService(name, namespace, nil)
		error := errors.New("fix")

		serviceSvc := automock.NewServiceSvc()
		serviceSvc.On("Find", name, namespace).Return(resource, nil).Once()
		defer serviceSvc.AssertExpectations(t)

		converter := automock.NewGqlServiceConverter()
		converter.On("ToGQL", resource).Return(nil, error).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(serviceSvc)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.DeleteServiceMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

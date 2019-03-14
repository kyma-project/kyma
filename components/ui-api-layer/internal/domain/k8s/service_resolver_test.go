package k8s_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	_assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

var (
	name      = "name"
	namespace = "namespace"
)

func TestServiceResolver_ServiceQuery(t *testing.T) {

	assert := _assert.New(t)

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

func TestserviceResolver_ServicesQuery(t *testing.T) {

	assert := _assert.New(t)

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
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlServiceConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ServicesQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var resources []*v1.Service
		var expected []gqlschema.Service

		resourceGetter := automock.NewServiceSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(resourceGetter)

		result, err := resolver.ServicesQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(expected, result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		expected := errors.New("test")
		var resources []*v1.Service
		resourceGetter := automock.NewServiceSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewServiceResolver(resourceGetter)

		result, err := resolver.ServicesQuery(nil, namespace, nil, nil)

		require.Error(t, err)
		assert.True(gqlerror.IsInternal(err))
		assert.Nil(result)
	})
}

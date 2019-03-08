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

func TestKServiceResolver_KServiceQuery(t *testing.T) {

	assert := _assert.New(t)

	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.KService{
			Name: "Test",
		}
		resource := &v1.Service{}
		resourceGetter := automock.NewKServiceSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlKServiceConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewKServiceResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.KServiceQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		resourceGetter := automock.NewKServiceSvc()
		resourceGetter.On("Find", name, namespace).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewKServiceResolver(resourceGetter)

		result, err := resolver.KServiceQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Nil(result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		expected := errors.New("test")
		resource := &v1.Service{}
		resourceGetter := automock.NewKServiceSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewKServiceResolver(resourceGetter)

		result, err := resolver.KServiceQuery(nil, name, namespace)

		assert.Error(err)
		assert.True(gqlerror.IsInternal(err))
		assert.Nil(result)
	})

}

func TestKserviceResolver_KServicesQuery(t *testing.T) {

	assert := _assert.New(t)

	t.Run("Success", func(t *testing.T) {

		resource := fixService(name, namespace, nil)
		resources := []*v1.Service{
			resource, resource,
		}
		expected := []gqlschema.KService{
			{
				Name: name,
			},
			{
				Name: name,
			},
		}

		resourceGetter := automock.NewKServiceSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlKServiceConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewKServiceResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.KServicesQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var resources []*v1.Service
		var expected []gqlschema.KService

		resourceGetter := automock.NewKServiceSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewKServiceResolver(resourceGetter)

		result, err := resolver.KServicesQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(expected, result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		expected := errors.New("test")
		var resources []*v1.Service
		resourceGetter := automock.NewKServiceSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewKServiceResolver(resourceGetter)

		result, err := resolver.KServicesQuery(nil, namespace, nil, nil)

		require.Error(t, err)
		assert.True(gqlerror.IsInternal(err))
		assert.Nil(result)
	})
}

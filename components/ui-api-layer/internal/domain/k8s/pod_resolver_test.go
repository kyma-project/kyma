package k8s_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestPodResolver_PodQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.Pod{
			Name: "Test",
		}
		name := "name"
		namespace := "namespace"
		resource := &v1.Pod{}
		resourceGetter := automock.NewPodLister()
		resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlPodConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.PodQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		namespace := "namespace"
		resourceGetter := automock.NewPodLister()
		resourceGetter.On("Find", name, namespace).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)

		result, err := resolver.PodQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"
		namespace := "namespace"
		resource := &v1.Pod{}
		resourceGetter := automock.NewPodLister()
		resourceGetter.On("Find", name, namespace).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)

		result, err := resolver.PodQuery(nil, name, namespace)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConverting", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"
		namespace := "namespace"
		resource := &v1.Pod{}
		resourceGetter := automock.NewPodLister()
		resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlPodConverter()
		converter.On("ToGQL", resource).Return(nil, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.PodQuery(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestPodResolver_PodsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "Test"
		namespace := "namespace"
		resource := fixPod(name, namespace, nil)
		resources := []*v1.Pod{
			resource, resource,
		}
		expected := []gqlschema.Pod{
			{
				Name: name,
			},
			{
				Name: name,
			},
		}

		resourceGetter := automock.NewPodLister()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlPodConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.PodsQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		namespace := "namespace"
		var resources []*v1.Pod
		var expected []gqlschema.Pod

		resourceGetter := automock.NewPodLister()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)

		result, err := resolver.PodsQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		namespace := "namespace"
		expected := errors.New("Test")
		var resources []*v1.Pod
		resourceGetter := automock.NewPodLister()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)

		result, err := resolver.PodsQuery(nil, namespace, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConverting", func(t *testing.T) {
		name := "Test"
		namespace := "namespace"
		resource := fixPod(name, namespace, nil)
		resources := []*v1.Pod{
			resource, resource,
		}
		expected := errors.New("Test")

		resourceGetter := automock.NewPodLister()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlPodConverter()
		converter.On("ToGQLs", resources).Return(nil, expected)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.PodsQuery(nil, namespace, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

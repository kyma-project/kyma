package k8s_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
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
		resourceGetter := automock.NewPodSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLPodConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)
		resolver.SetPodConverter(converter)

		result, err := resolver.PodQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		namespace := "namespace"
		resourceGetter := automock.NewPodSvc()
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
		resourceGetter := automock.NewPodSvc()
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
		resourceGetter := automock.NewPodSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLPodConverter()
		converter.On("ToGQL", resource).Return(nil, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)
		resolver.SetPodConverter(converter)

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
		expected := []*gqlschema.Pod{
			{
				Name: name,
			},
			{
				Name: name,
			},
		}

		resourceGetter := automock.NewPodSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLPodConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)
		resolver.SetPodConverter(converter)

		result, err := resolver.PodsQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		namespace := "namespace"
		var resources []*v1.Pod
		var expected []*gqlschema.Pod

		resourceGetter := automock.NewPodSvc()
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
		resourceGetter := automock.NewPodSvc()
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

		resourceGetter := automock.NewPodSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLPodConverter()
		converter.On("ToGQLs", resources).Return(nil, expected)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(resourceGetter)
		resolver.SetPodConverter(converter)

		result, err := resolver.PodsQuery(nil, namespace, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestPodResolver_PodEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewPodSvc()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := k8s.NewPodResolver(svc)

		_, err := resolver.PodEventSubscription(ctx, "test")

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewPodSvc()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := k8s.NewPodResolver(svc)

		channel, err := resolver.PodEventSubscription(ctx, "test")
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}

func TestPodResolver_UpdatePodMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedPodFix := fixPod(name, namespace, map[string]string{
			"test": "test",
		})
		updatedGQLPodFix := &gqlschema.Pod{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"test": "test",
			},
		}
		gqlJSONFix := gqlschema.JSON{}

		podSvc := automock.NewPodSvc()
		podSvc.On("Update", name, namespace, *updatedPodFix).Return(updatedPodFix, nil).Once()
		defer podSvc.AssertExpectations(t)

		converter := automock.NewGQLPodConverter()
		converter.On("GQLJSONToPod", gqlJSONFix).Return(*updatedPodFix, nil).Once()
		converter.On("ToGQL", updatedPodFix).Return(updatedGQLPodFix, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(podSvc)
		resolver.SetPodConverter(converter)

		result, err := resolver.UpdatePodMutation(nil, name, namespace, gqlJSONFix)

		require.NoError(t, err)
		assert.Equal(t, updatedGQLPodFix, result)
	})

	t.Run("ErrorConvertingToPod", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		podSvc := automock.NewPodSvc()

		converter := automock.NewGQLPodConverter()
		converter.On("GQLJSONToPod", gqlJSONFix).Return(v1.Pod{}, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(podSvc)
		resolver.SetPodConverter(converter)

		result, err := resolver.UpdatePodMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorUpdating", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedPodFix := fixPod(name, namespace, map[string]string{
			"test": "test",
		})
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		podSvc := automock.NewPodSvc()
		podSvc.On("Update", name, namespace, *updatedPodFix).Return(nil, expected).Once()
		defer podSvc.AssertExpectations(t)

		converter := automock.NewGQLPodConverter()
		converter.On("GQLJSONToPod", gqlJSONFix).Return(*updatedPodFix, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(podSvc)
		resolver.SetPodConverter(converter)

		result, err := resolver.UpdatePodMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ErrorConvertingToGQL", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedPodFix := fixPod(name, namespace, map[string]string{
			"test": "test",
		})
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		podSvc := automock.NewPodSvc()
		podSvc.On("Update", name, namespace, *updatedPodFix).Return(updatedPodFix, nil).Once()
		defer podSvc.AssertExpectations(t)

		converter := automock.NewGQLPodConverter()
		converter.On("GQLJSONToPod", gqlJSONFix).Return(*updatedPodFix, nil).Once()
		converter.On("ToGQL", updatedPodFix).Return(nil, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(podSvc)
		resolver.SetPodConverter(converter)

		result, err := resolver.UpdatePodMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestPodResolver_DeletePodMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixPod(name, namespace, nil)
		expected := &gqlschema.Pod{
			Name:      name,
			Namespace: namespace,
		}

		podSvc := automock.NewPodSvc()
		podSvc.On("Find", name, namespace).Return(resource, nil).Once()
		podSvc.On("Delete", name, namespace).Return(nil).Once()
		defer podSvc.AssertExpectations(t)

		converter := automock.NewGQLPodConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(podSvc)
		resolver.SetPodConverter(converter)

		result, err := resolver.DeletePodMutation(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		expected := errors.New("fix")

		podSvc := automock.NewPodSvc()
		podSvc.On("Find", name, namespace).Return(nil, expected).Once()
		defer podSvc.AssertExpectations(t)

		resolver := k8s.NewPodResolver(podSvc)

		result, err := resolver.DeletePodMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorDeleting", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixPod(name, namespace, nil)
		expected := errors.New("fix")

		podSvc := automock.NewPodSvc()
		podSvc.On("Find", name, namespace).Return(resource, nil).Once()
		podSvc.On("Delete", name, namespace).Return(expected).Once()
		defer podSvc.AssertExpectations(t)

		resolver := k8s.NewPodResolver(podSvc)

		result, err := resolver.DeletePodMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConverting", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixPod(name, namespace, nil)
		error := errors.New("fix")

		podSvc := automock.NewPodSvc()
		podSvc.On("Find", name, namespace).Return(resource, nil).Once()
		defer podSvc.AssertExpectations(t)

		converter := automock.NewGQLPodConverter()
		converter.On("ToGQL", resource).Return(nil, error).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewPodResolver(podSvc)
		resolver.SetPodConverter(converter)

		result, err := resolver.DeletePodMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

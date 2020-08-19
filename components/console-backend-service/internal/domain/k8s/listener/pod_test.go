package listener_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestPodListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlPod := new(gqlschema.Pod)
		pod := new(v1.Pod)
		converter := automock.NewGQLPodConverter()

		channel := make(chan *gqlschema.PodEvent, 1)
		defer close(channel)
		converter.On("ToGQL", pod).Return(gqlPod, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewPod(channel, filterPodTrue, converter)

		// when
		podListener.OnAdd(pod)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, gqlPod, result.Pod)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		podListener := listener.NewPod(nil, filterPodFalse, nil)

		// when
		podListener.OnAdd(new(v1.Pod))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		podListener := listener.NewPod(nil, filterPodTrue, nil)

		// when
		podListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		pod := new(v1.Pod)
		converter := automock.NewGQLPodConverter()

		converter.On("ToGQL", pod).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewPod(nil, filterPodTrue, converter)

		// when
		podListener.OnAdd(pod)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		podListener := listener.NewPod(nil, filterPodTrue, nil)

		// when
		podListener.OnAdd(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		pod := new(v1.Pod)
		converter := automock.NewGQLPodConverter()

		converter.On("ToGQL", pod).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewPod(nil, filterPodTrue, converter)

		// when
		podListener.OnAdd(pod)
	})
}

func TestPodListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlPod := new(gqlschema.Pod)
		pod := new(v1.Pod)
		converter := automock.NewGQLPodConverter()

		channel := make(chan *gqlschema.PodEvent, 1)
		defer close(channel)
		converter.On("ToGQL", pod).Return(gqlPod, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewPod(channel, filterPodTrue, converter)

		// when
		podListener.OnDelete(pod)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, gqlPod, result.Pod)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		podListener := listener.NewPod(nil, filterPodFalse, nil)

		// when
		podListener.OnDelete(new(v1.Pod))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		podListener := listener.NewPod(nil, filterPodTrue, nil)

		// when
		podListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		pod := new(v1.Pod)
		converter := automock.NewGQLPodConverter()

		converter.On("ToGQL", pod).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewPod(nil, filterPodTrue, converter)

		// when
		podListener.OnDelete(pod)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		podListener := listener.NewPod(nil, filterPodTrue, nil)

		// when
		podListener.OnDelete(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		pod := new(v1.Pod)
		converter := automock.NewGQLPodConverter()

		converter.On("ToGQL", pod).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewPod(nil, filterPodTrue, converter)

		// when
		podListener.OnDelete(pod)
	})
}

func TestPodListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlPod := new(gqlschema.Pod)
		pod := new(v1.Pod)
		converter := automock.NewGQLPodConverter()

		channel := make(chan *gqlschema.PodEvent, 1)
		defer close(channel)
		converter.On("ToGQL", pod).Return(gqlPod, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewPod(channel, filterPodTrue, converter)

		// when
		podListener.OnUpdate(pod, pod)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, gqlPod, result.Pod)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		podListener := listener.NewPod(nil, filterPodFalse, nil)

		// when
		podListener.OnUpdate(new(v1.Pod), new(v1.Pod))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		podListener := listener.NewPod(nil, filterPodTrue, nil)

		// when
		podListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		pod := new(v1.Pod)
		converter := automock.NewGQLPodConverter()

		converter.On("ToGQL", pod).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewPod(nil, filterPodTrue, converter)

		// when
		podListener.OnUpdate(nil, pod)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		podListener := listener.NewPod(nil, filterPodTrue, nil)

		// when
		podListener.OnUpdate(new(struct{}), new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		pod := new(v1.Pod)
		converter := automock.NewGQLPodConverter()

		converter.On("ToGQL", pod).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewPod(nil, filterPodTrue, converter)

		// when
		podListener.OnUpdate(nil, pod)
	})
}

func filterPodTrue(o *v1.Pod) bool {
	return true
}

func filterPodFalse(o *v1.Pod) bool {
	return false
}

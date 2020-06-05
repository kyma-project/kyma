package listener_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNamespaceListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlNamespace := new(gqlschema.Namespace)
		namespace := new(v1.Namespace)
		converter := automock.NewNamespaceConverter()

		channel := make(chan gqlschema.NamespaceEvent, 1)
		defer close(channel)
		converter.On("ToGQL", namespace).Return(gqlNamespace).Once()
		defer converter.AssertExpectations(t)
		namespaceListener := listener.NewNamespace(channel, filterNamespaceTrue, converter, []string{})

		// when
		namespaceListener.OnAdd(namespace)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlNamespace, result.Namespace)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		namespaceListener := listener.NewNamespace(nil, filterNamespaceFalse, nil, []string{})

		// when
		namespaceListener.OnAdd(new(v1.Namespace))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		namespaceListener := listener.NewNamespace(nil, filterNamespaceTrue, nil, []string{})

		// when
		namespaceListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		namespace := new(v1.Namespace)
		converter := automock.NewNamespaceConverter()

		converter.On("ToGQL", namespace).Return(nil).Once()
		defer converter.AssertExpectations(t)
		namespaceListener := listener.NewNamespace(nil, filterNamespaceTrue, converter, []string{})

		// when
		namespaceListener.OnAdd(namespace)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		namespaceListener := listener.NewNamespace(nil, filterNamespaceTrue, nil, []string{})

		// when
		namespaceListener.OnAdd(new(struct{}))
	})
}

func TestNamespaceListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlNamespace := new(gqlschema.Namespace)
		namespace := new(v1.Namespace)
		converter := automock.NewNamespaceConverter()

		channel := make(chan gqlschema.NamespaceEvent, 1)
		defer close(channel)
		converter.On("ToGQL", namespace).Return(gqlNamespace).Once()
		defer converter.AssertExpectations(t)
		namespaceListener := listener.NewNamespace(channel, filterNamespaceTrue, converter, []string{})

		// when
		namespaceListener.OnDelete(namespace)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlNamespace, result.Namespace)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		namespaceListener := listener.NewNamespace(nil, filterNamespaceFalse, nil, []string{})

		// when
		namespaceListener.OnDelete(new(v1.Namespace))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		namespaceListener := listener.NewNamespace(nil, filterNamespaceTrue, nil, []string{})

		// when
		namespaceListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		namespace := new(v1.Namespace)
		converter := automock.NewNamespaceConverter()

		converter.On("ToGQL", namespace).Return(nil).Once()
		defer converter.AssertExpectations(t)
		namespaceListener := listener.NewNamespace(nil, filterNamespaceTrue, converter, []string{})

		// when
		namespaceListener.OnDelete(namespace)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		namespaceListener := listener.NewNamespace(nil, filterNamespaceTrue, nil, []string{})

		// when
		namespaceListener.OnDelete(new(struct{}))
	})
}

func TestNamespaceListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlNamespace := new(gqlschema.Namespace)
		namespace := new(v1.Namespace)
		converter := automock.NewNamespaceConverter()

		channel := make(chan gqlschema.NamespaceEvent, 1)
		defer close(channel)
		converter.On("ToGQL", namespace).Return(gqlNamespace).Once()
		defer converter.AssertExpectations(t)
		namespaceListener := listener.NewNamespace(channel, filterNamespaceTrue, converter, []string{})

		// when
		namespaceListener.OnUpdate(namespace, namespace)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlNamespace, result.Namespace)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		namespaceListener := listener.NewNamespace(nil, filterNamespaceFalse, nil, []string{})

		// when
		namespaceListener.OnUpdate(new(v1.Namespace), new(v1.Namespace))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		namespaceListener := listener.NewNamespace(nil, filterNamespaceTrue, nil, []string{})

		// when
		namespaceListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		namespace := new(v1.Namespace)
		converter := automock.NewNamespaceConverter()

		converter.On("ToGQL", namespace).Return(nil).Once()
		defer converter.AssertExpectations(t)
		namespaceListener := listener.NewNamespace(nil, filterNamespaceTrue, converter, []string{})

		// when
		namespaceListener.OnUpdate(nil, namespace)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		namespaceListener := listener.NewNamespace(nil, filterNamespaceTrue, nil, []string{})

		// when
		namespaceListener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterNamespaceTrue(_ *v1.Namespace) bool {
	return true
}

func filterNamespaceFalse(_ *v1.Namespace) bool {
	return false
}

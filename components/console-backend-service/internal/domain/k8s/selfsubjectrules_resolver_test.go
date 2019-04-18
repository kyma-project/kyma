package k8s_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	authv1 "k8s.io/api/authorization/v1"
)

func TestSelfSubjectRulesQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.SelfSubjectRules{}

		resource := &authv1.SelfSubjectRulesReview{}

		resourceGetter := automock.NewSelfSubjectRulesSvc()
		// resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		// defer resourceGetter.AssertExpectations(t)

		// converter := automock.NewGqlReplicaSetConverter()
		// converter.On("ToGQL", resource).Return(expected, nil).Once()
		// defer converter.AssertExpectations(t)

		// resolver := k8s.NewReplicaSetResolver(resourceGetter)
		// resolver.SetInstanceConverter(converter)

		// result, err := resolver.ReplicaSetQuery(nil, name, namespace)

		// require.NoError(t, err)
		// assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		namespace := "namespace"
		resourceGetter := automock.NewReplicaSetSvc()
		resourceGetter.On("Find", name, namespace).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(resourceGetter)

		result, err := resolver.ReplicaSetQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"
		namespace := "namespace"
		resource := &apps.ReplicaSet{}
		resourceGetter := automock.NewReplicaSetSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(resourceGetter)

		result, err := resolver.ReplicaSetQuery(nil, name, namespace)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConverting", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"
		namespace := "namespace"
		resource := &apps.ReplicaSet{}
		resourceGetter := automock.NewReplicaSetSvc()
		resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlReplicaSetConverter()
		converter.On("ToGQL", resource).Return(nil, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ReplicaSetQuery(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestReplicaSetResolver_ReplicaSetsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "Test"
		namespace := "namespace"
		resource := fixReplicaSet(name, namespace, map[string]string{
			"test": "test",
		})
		resources := []*apps.ReplicaSet{
			resource, resource,
		}
		expected := []gqlschema.ReplicaSet{
			{
				Name: name,
			},
			{
				Name: name,
			},
		}

		resourceGetter := automock.NewReplicaSetSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlReplicaSetConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ReplicaSetsQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		namespace := "namespace"
		var resources []*apps.ReplicaSet
		var expected []gqlschema.ReplicaSet

		resourceGetter := automock.NewReplicaSetSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(resourceGetter)

		result, err := resolver.ReplicaSetsQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorGetting", func(t *testing.T) {
		namespace := "namespace"
		expected := errors.New("Test")
		var resources []*apps.ReplicaSet
		resourceGetter := automock.NewReplicaSetSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(resourceGetter)

		result, err := resolver.ReplicaSetsQuery(nil, namespace, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConverting", func(t *testing.T) {
		name := "Test"
		namespace := "namespace"
		resource := fixReplicaSet(name, namespace, map[string]string{
			"test": "test",
		})
		resources := []*apps.ReplicaSet{
			resource, resource,
		}
		expected := errors.New("Test")

		resourceGetter := automock.NewReplicaSetSvc()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGqlReplicaSetConverter()
		converter.On("ToGQLs", resources).Return(nil, expected)
		defer converter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ReplicaSetsQuery(nil, namespace, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestReplicaSetResolver_UpdateReplicaSetMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedReplicaSetFix := fixReplicaSet(name, namespace, map[string]string{
			"test": "test",
		})
		updatedGQLReplicaSetFix := &gqlschema.ReplicaSet{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"test": "test",
			},
		}
		gqlJSONFix := gqlschema.JSON{}

		replicaSetSvc := automock.NewReplicaSetSvc()
		replicaSetSvc.On("Update", name, namespace, *updatedReplicaSetFix).Return(updatedReplicaSetFix, nil).Once()
		defer replicaSetSvc.AssertExpectations(t)

		converter := automock.NewGqlReplicaSetConverter()
		converter.On("GQLJSONToReplicaSet", gqlJSONFix).Return(*updatedReplicaSetFix, nil).Once()
		converter.On("ToGQL", updatedReplicaSetFix).Return(updatedGQLReplicaSetFix, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(replicaSetSvc)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.UpdateReplicaSetMutation(nil, name, namespace, gqlJSONFix)

		require.NoError(t, err)
		assert.Equal(t, updatedGQLReplicaSetFix, result)
	})

	t.Run("ErrorConvertingToReplicaSet", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		replicaSetSvc := automock.NewReplicaSetSvc()

		converter := automock.NewGqlReplicaSetConverter()
		converter.On("GQLJSONToReplicaSet", gqlJSONFix).Return(apps.ReplicaSet{}, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(replicaSetSvc)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.UpdateReplicaSetMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorUpdating", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedReplicaSetFix := fixReplicaSet(name, namespace, map[string]string{
			"test": "test",
		})
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		replicaSetSvc := automock.NewReplicaSetSvc()
		replicaSetSvc.On("Update", name, namespace, *updatedReplicaSetFix).Return(nil, expected).Once()
		defer replicaSetSvc.AssertExpectations(t)

		converter := automock.NewGqlReplicaSetConverter()
		converter.On("GQLJSONToReplicaSet", gqlJSONFix).Return(*updatedReplicaSetFix, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(replicaSetSvc)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.UpdateReplicaSetMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ErrorConvertingToGQL", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		updatedReplicaSetFix := fixReplicaSet(name, namespace, map[string]string{
			"test": "test",
		})
		gqlJSONFix := gqlschema.JSON{}
		expected := errors.New("fix")

		replicaSetSvc := automock.NewReplicaSetSvc()
		replicaSetSvc.On("Update", name, namespace, *updatedReplicaSetFix).Return(updatedReplicaSetFix, nil).Once()
		defer replicaSetSvc.AssertExpectations(t)

		converter := automock.NewGqlReplicaSetConverter()
		converter.On("GQLJSONToReplicaSet", gqlJSONFix).Return(*updatedReplicaSetFix, nil).Once()
		converter.On("ToGQL", updatedReplicaSetFix).Return(nil, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(replicaSetSvc)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.UpdateReplicaSetMutation(nil, name, namespace, gqlJSONFix)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestReplicaSetResolver_DeleteReplicaSetMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixReplicaSet(name, namespace, map[string]string{
			"test": "test",
		})
		expected := &gqlschema.ReplicaSet{
			Name:      name,
			Namespace: namespace,
		}

		replicaSetSvc := automock.NewReplicaSetSvc()
		replicaSetSvc.On("Find", name, namespace).Return(resource, nil).Once()
		replicaSetSvc.On("Delete", name, namespace).Return(nil).Once()
		defer replicaSetSvc.AssertExpectations(t)

		converter := automock.NewGqlReplicaSetConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(replicaSetSvc)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.DeleteReplicaSetMutation(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		expected := errors.New("fix")

		replicaSetSvc := automock.NewReplicaSetSvc()
		replicaSetSvc.On("Find", name, namespace).Return(nil, expected).Once()
		defer replicaSetSvc.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(replicaSetSvc)

		result, err := resolver.DeleteReplicaSetMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorDeleting", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixReplicaSet(name, namespace, map[string]string{
			"test": "test",
		})
		expected := errors.New("fix")

		replicaSetSvc := automock.NewReplicaSetSvc()
		replicaSetSvc.On("Find", name, namespace).Return(resource, nil).Once()
		replicaSetSvc.On("Delete", name, namespace).Return(expected).Once()
		defer replicaSetSvc.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(replicaSetSvc)

		result, err := resolver.DeleteReplicaSetMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConverting", func(t *testing.T) {
		name := "exampleName"
		namespace := "exampleNamespace"
		resource := fixReplicaSet(name, namespace, map[string]string{
			"test": "test",
		})
		error := errors.New("fix")

		replicaSetSvc := automock.NewReplicaSetSvc()
		replicaSetSvc.On("Find", name, namespace).Return(resource, nil).Once()
		defer replicaSetSvc.AssertExpectations(t)

		converter := automock.NewGqlReplicaSetConverter()
		converter.On("ToGQL", resource).Return(nil, error).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewReplicaSetResolver(replicaSetSvc)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.DeleteReplicaSetMutation(nil, name, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

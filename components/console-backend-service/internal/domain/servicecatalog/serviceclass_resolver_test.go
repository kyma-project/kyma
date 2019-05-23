package servicecatalog_test

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/automock"
	cmsMock "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClassResolver_ServiceClassQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "name"
		ns := "ns"
		expected := &gqlschema.ServiceClass{
			Name: "Test",
		}
		resource := &v1beta1.ServiceClass{}
		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("Find", name, ns).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLServiceClassConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)
		resolver.SetClassConverter(converter)

		result, err := resolver.ServiceClassQuery(nil, name, ns)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		ns := "ns"
		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("Find", name, ns).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)

		result, err := resolver.ServiceClassQuery(nil, name, ns)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		name := "name"
		ns := "ns"
		expected := errors.New("Test")
		resource := &v1beta1.ServiceClass{}
		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("Find", name, ns).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)

		result, err := resolver.ServiceClassQuery(nil, name, ns)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClassResolver_ServiceClassesQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ns := "ns"
		resource :=
			&v1beta1.ServiceClass{
				ObjectMeta: v1.ObjectMeta{
					Name: "test",
				},
			}
		resources := []*v1beta1.ServiceClass{
			resource, resource,
		}
		expected := []gqlschema.ServiceClass{
			{
				Name: "Test",
			}, {
				Name: "Test",
			},
		}

		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("List", ns, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLServiceClassConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)
		resolver.SetClassConverter(converter)

		result, err := resolver.ServiceClassesQuery(nil, ns, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		ns := "ns"
		var resources []*v1beta1.ServiceClass

		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("List", ns, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)
		var expected []gqlschema.ServiceClass

		result, err := resolver.ServiceClassesQuery(nil, ns, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		ns := "ns"
		expected := errors.New("Test")

		var resources []*v1beta1.ServiceClass

		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("List", ns, pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)

		_, err := resolver.ServiceClassesQuery(nil, ns, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestClassResolver_ServiceClassInstancesField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ns := "ns"
		name := "name"
		externalName := "externalName"
		resources := []*v1beta1.ServiceInstance{
			fixServiceInstance("foo", "ns"),
			fixServiceInstance("bar", "ns"),
		}
		expected := []gqlschema.ServiceInstance{
			{Name: "foo", Namespace: ns},
			{Name: "bar", Namespace: ns},
		}

		resourceGetter := automock.NewInstanceListerByServiceClass()
		resourceGetter.On("ListForServiceClass", name, externalName, ns).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
			Namespace:    ns,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, resourceGetter, nil)

		result, err := resolver.ServiceClassInstancesField(nil, &parentObj)

		require.NoError(t, err)
		assert.Len(t, result, len(expected))
		for idx, e := range expected {
			assert.Equal(t, e.Name, result[idx].Name)
			assert.Equal(t, e.Namespace, result[idx].Namespace)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		ns := "ns"
		name := "name"
		externalName := "externalName"
		var expected []gqlschema.ServiceInstance
		resourceGetter := automock.NewInstanceListerByServiceClass()
		resourceGetter.On("ListForServiceClass", name, externalName, ns).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := &gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
			Namespace:    ns,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, resourceGetter, nil)

		result, err := resolver.ServiceClassInstancesField(nil, parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		ns := "ns"
		expectedErr := errors.New("Test")
		name := "name"
		externalName := "externalName"
		resourceGetter := automock.NewInstanceListerByServiceClass()
		resourceGetter.On("ListForServiceClass", name, externalName, ns).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
			Namespace:    ns,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, resourceGetter, nil)

		_, err := resolver.ServiceClassInstancesField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestClassResolver_ServiceClassActivatedField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ns := "ns"
		name := "name"
		externalName := "externalName"
		resources := []*v1beta1.ServiceInstance{{}, {}}
		resourceGetter := automock.NewInstanceListerByServiceClass()
		resourceGetter.On("ListForServiceClass", name, externalName, ns).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
			Namespace:    ns,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, resourceGetter, nil)

		result, err := resolver.ServiceClassActivatedField(nil, &parentObj)

		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		ns := "ns"
		name := "name"
		externalName := "externalName"
		resourceGetter := automock.NewInstanceListerByServiceClass()
		resourceGetter.On("ListForServiceClass", name, externalName, ns).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := &gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
			Namespace:    ns,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, resourceGetter, nil)

		result, err := resolver.ServiceClassActivatedField(nil, parentObj)

		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		ns := "ns"
		expectedErr := errors.New("Test")
		name := "name"
		externalName := "externalName"
		resourceGetter := automock.NewInstanceListerByServiceClass()
		resourceGetter.On("ListForServiceClass", name, externalName, ns).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
			Namespace:    ns,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, resourceGetter, nil)

		_, err := resolver.ServiceClassActivatedField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestClassResolver_ServiceClassPlansField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ns := "ns"
		expectedSingleObj := gqlschema.ServicePlan{
			Name: "Test",
		}
		expected := []gqlschema.ServicePlan{
			expectedSingleObj,
			expectedSingleObj,
		}

		name := "name"
		resource := v1beta1.ServicePlan{}
		resources := []*v1beta1.ServicePlan{
			&resource,
			&resource,
		}
		resourceGetter := automock.NewServicePlanLister()
		resourceGetter.On("ListForServiceClass", name, ns).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLServicePlanConverter()
		converter.On("ToGQLs", resources).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:      name,
			Namespace: ns,
		}
		resolver := servicecatalog.NewServiceClassResolver(nil, resourceGetter, nil, nil)
		resolver.SetPlanConverter(converter)

		result, err := resolver.ServiceClassPlansField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		ns := "ns"
		name := "name"
		resourceGetter := automock.NewServicePlanLister()
		resourceGetter.On("ListForServiceClass", name, ns).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:      name,
			Namespace: ns,
		}
		resolver := servicecatalog.NewServiceClassResolver(nil, resourceGetter, nil, nil)

		result, err := resolver.ServiceClassPlansField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		ns := "ns"
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := automock.NewServicePlanLister()
		resourceGetter.On("ListForServiceClass", name, ns).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:      name,
			Namespace: ns,
		}
		resolver := servicecatalog.NewServiceClassResolver(nil, resourceGetter, nil, nil)

		result, err := resolver.ServiceClassPlansField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClassResolver_ServiceClassClusterDocsTopicsField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "name"
		namespace := "namespace"
		resources := &v1alpha1.ClusterDocsTopic{
			ObjectMeta: v1.ObjectMeta{
				Name: name,
			},
		}
		expected := &gqlschema.ClusterDocsTopic{
			Name: name,
		}

		resourceGetter := new(cmsMock.ClusterDocsTopicGetter)
		resourceGetter.On("Find", name).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := new(cmsMock.GqlClusterDocsTopicConverter)
		converter.On("ToGQL", resources).Return(expected, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(cmsMock.CmsRetriever)
		retriever.On("ClusterDocsTopic").Return(resourceGetter)
		retriever.On("ClusterDocsTopicConverter").Return(converter)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: name,
			Namespace:    namespace,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassClusterDocsTopicField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		namespace := "namespace"

		resourceGetter := new(cmsMock.ClusterDocsTopicGetter)
		resourceGetter.On("Find", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := new(cmsMock.GqlClusterDocsTopicConverter)
		converter.On("ToGQL", (*v1alpha1.ClusterDocsTopic)(nil)).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(cmsMock.CmsRetriever)
		retriever.On("ClusterDocsTopic").Return(resourceGetter)
		retriever.On("ClusterDocsTopicConverter").Return(converter)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: name,
			Namespace:    namespace,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassClusterDocsTopicField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		namespace := "namespace"

		resourceGetter := new(cmsMock.ClusterDocsTopicGetter)
		resourceGetter.On("Find", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(cmsMock.CmsRetriever)
		retriever.On("ClusterDocsTopic").Return(resourceGetter)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: name,
			Namespace:    namespace,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassClusterDocsTopicField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClassResolver_ServiceClassDocsTopicsField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "name"
		namespace := "namespace"
		resources := &v1alpha1.DocsTopic{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		expected := &gqlschema.DocsTopic{
			Name:      name,
			Namespace: namespace,
		}

		resourceGetter := new(cmsMock.DocsTopicGetter)
		resourceGetter.On("Find", namespace, name).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := new(cmsMock.GqlDocsTopicConverter)
		converter.On("ToGQL", resources).Return(expected, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(cmsMock.CmsRetriever)
		retriever.On("DocsTopic").Return(resourceGetter)
		retriever.On("DocsTopicConverter").Return(converter)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: name,
			Namespace:    namespace,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassDocsTopicField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		namespace := "namespace"

		resourceGetter := new(cmsMock.DocsTopicGetter)
		resourceGetter.On("Find", namespace, name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := new(cmsMock.GqlDocsTopicConverter)
		converter.On("ToGQL", (*v1alpha1.DocsTopic)(nil)).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(cmsMock.CmsRetriever)
		retriever.On("DocsTopic").Return(resourceGetter)
		retriever.On("DocsTopicConverter").Return(converter)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: name,
			Namespace:    namespace,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassDocsTopicField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		namespace := "namespace"

		resourceGetter := new(cmsMock.DocsTopicGetter)
		resourceGetter.On("Find", namespace, name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(cmsMock.CmsRetriever)
		retriever.On("DocsTopic").Return(resourceGetter)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: name,
			Namespace:    namespace,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassDocsTopicField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

package servicecatalog_test

import (
	"fmt"
	"testing"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/automock"
	mock "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterServiceClassResolver_ClusterServiceClassQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.ClusterServiceClass{
			Name: "Test",
		}
		name := "name"
		resource := &v1beta1.ClusterServiceClass{}
		resourceGetter := automock.NewClusterServiceClassListGetter()
		resourceGetter.On("Find", name).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLClusterServiceClassConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil)
		resolver.SetClassConverter(converter)

		result, err := resolver.ClusterServiceClassQuery(nil, name)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := automock.NewClusterServiceClassListGetter()
		resourceGetter.On("Find", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil)

		result, err := resolver.ClusterServiceClassQuery(nil, name)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"
		resource := &v1beta1.ClusterServiceClass{}
		resourceGetter := automock.NewClusterServiceClassListGetter()
		resourceGetter.On("Find", name).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil)

		result, err := resolver.ClusterServiceClassQuery(nil, name)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClusterServiceClassResolver_ClusterServiceClassesQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resource :=
			&v1beta1.ClusterServiceClass{
				ObjectMeta: v1.ObjectMeta{
					Name: "test",
				},
			}
		resources := []*v1beta1.ClusterServiceClass{
			resource, resource,
		}
		expected := []gqlschema.ClusterServiceClass{
			{
				Name: "Test",
			}, {
				Name: "Test",
			},
		}

		resourceGetter := automock.NewClusterServiceClassListGetter()
		resourceGetter.On("List", pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLClusterServiceClassConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil)
		resolver.SetClassConverter(converter)

		result, err := resolver.ClusterServiceClassesQuery(nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var resources []*v1beta1.ClusterServiceClass

		resourceGetter := automock.NewClusterServiceClassListGetter()
		resourceGetter.On("List", pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil)
		var expected []gqlschema.ClusterServiceClass

		result, err := resolver.ClusterServiceClassesQuery(nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")

		var resources []*v1beta1.ClusterServiceClass

		resourceGetter := automock.NewClusterServiceClassListGetter()
		resourceGetter.On("List", pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil)

		_, err := resolver.ClusterServiceClassesQuery(nil, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestClusterServiceClassResolver_ClusterServiceClassInstancesField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {

		testNs := "test"
		for testNo, testCase := range []struct {
			Namespace *string
		}{
			{Namespace: nil},
			{Namespace: &testNs},
		} {
			t.Run(fmt.Sprintf("Test Case %d", testNo), func(t *testing.T) {
				name := "name"
				ns := "ns"
				externalName := "externalName"
				resources := []*v1beta1.ServiceInstance{
					fixServiceInstance("foo", "ns"),
					fixServiceInstance("bar", "ns"),
				}
				expected := []gqlschema.ServiceInstance{
					{Name: "foo", Namespace: ns},
					{Name: "bar", Namespace: ns},
				}

				resourceGetter := automock.NewInstanceListerByClusterServiceClass()
				resourceGetter.On("ListForClusterServiceClass", name, externalName, testCase.Namespace).Return(resources, nil).Once()
				defer resourceGetter.AssertExpectations(t)

				parentObj := gqlschema.ClusterServiceClass{
					Name:         name,
					ExternalName: externalName,
				}

				resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, resourceGetter, nil, nil)

				result, err := resolver.ClusterServiceClassInstancesField(nil, &parentObj, testCase.Namespace)

				require.NoError(t, err)
				assert.Len(t, result, len(expected))
				for idx, e := range expected {
					assert.Equal(t, e.Name, result[idx].Name)
					assert.Equal(t, e.Namespace, result[idx].Namespace)
				}
			})
		}
	})

	t.Run("NotFound", func(t *testing.T) {

		testNs := "test"
		for testNo, testCase := range []struct {
			Namespace *string
		}{
			{Namespace: nil},
			{Namespace: &testNs},
		} {
			t.Run(fmt.Sprintf("Test Case %d", testNo), func(t *testing.T) {
				name := "name"
				externalName := "externalName"
				var expected []gqlschema.ServiceInstance
				resourceGetter := automock.NewInstanceListerByClusterServiceClass()
				resourceGetter.On("ListForClusterServiceClass", name, externalName, testCase.Namespace).Return(nil, nil).Once()
				defer resourceGetter.AssertExpectations(t)

				parentObj := &gqlschema.ClusterServiceClass{
					Name:         name,
					ExternalName: externalName,
				}

				resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, resourceGetter, nil, nil)

				result, err := resolver.ClusterServiceClassInstancesField(nil, parentObj, testCase.Namespace)

				require.NoError(t, err)
				assert.Equal(t, expected, result)
			})
		}
	})

	t.Run("Error", func(t *testing.T) {

		testNs := "test"
		for testNo, testCase := range []struct {
			Namespace *string
		}{
			{Namespace: nil},
			{Namespace: &testNs},
		} {
			t.Run(fmt.Sprintf("Test Case %d", testNo), func(t *testing.T) {
				expectedErr := errors.New("Test")
				name := "name"
				externalName := "externalName"
				resourceGetter := automock.NewInstanceListerByClusterServiceClass()
				resourceGetter.On("ListForClusterServiceClass", name, externalName, testCase.Namespace).Return(nil, expectedErr).Once()
				defer resourceGetter.AssertExpectations(t)

				parentObj := gqlschema.ClusterServiceClass{
					Name:         name,
					ExternalName: externalName,
				}

				resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, resourceGetter, nil, nil)

				_, err := resolver.ClusterServiceClassInstancesField(nil, &parentObj, testCase.Namespace)

				assert.Error(t, err)
				assert.True(t, gqlerror.IsInternal(err))
			})
		}

	})
}

func TestClusterServiceClassResolver_ClusterServiceClassActivatedField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		testNs := "test"
		for testNo, testCase := range []struct {
			Namespace *string
		}{
			{Namespace: nil},
			{Namespace: &testNs},
		} {
			t.Run(fmt.Sprintf("Test Case %d", testNo), func(t *testing.T) {
				expected := true
				name := "name"
				externalName := "externalName"
				resources := []*v1beta1.ServiceInstance{{}, {}}
				resourceGetter := automock.NewInstanceListerByClusterServiceClass()
				resourceGetter.On("ListForClusterServiceClass", name, externalName, testCase.Namespace).Return(resources, nil).Once()
				defer resourceGetter.AssertExpectations(t)

				parentObj := gqlschema.ClusterServiceClass{
					Name:         name,
					ExternalName: externalName,
				}

				resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, resourceGetter, nil, nil)

				result, err := resolver.ClusterServiceClassActivatedField(nil, &parentObj, testCase.Namespace)

				require.NoError(t, err)
				assert.Equal(t, expected, result)
			})
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		testNs := "test"
		for testNo, testCase := range []struct {
			Namespace *string
		}{
			{Namespace: nil},
			{Namespace: &testNs},
		} {
			t.Run(fmt.Sprintf("Test Case %d", testNo), func(t *testing.T) {
				name := "name"
				externalName := "externalName"
				resourceGetter := automock.NewInstanceListerByClusterServiceClass()
				resourceGetter.On("ListForClusterServiceClass", name, externalName, testCase.Namespace).Return(nil, nil).Once()
				defer resourceGetter.AssertExpectations(t)

				parentObj := &gqlschema.ClusterServiceClass{
					Name:         name,
					ExternalName: externalName,
				}

				resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, resourceGetter, nil, nil)

				result, err := resolver.ClusterServiceClassActivatedField(nil, parentObj, testCase.Namespace)

				require.NoError(t, err)
				assert.False(t, result)
			})
		}
	})

	t.Run("Error", func(t *testing.T) {
		testNs := "test"
		for testNo, testCase := range []struct {
			Namespace *string
		}{
			{Namespace: nil},
			{Namespace: &testNs},
		} {
			t.Run(fmt.Sprintf("Test Case %d", testNo), func(t *testing.T) {

				expectedErr := errors.New("Test")
				name := "name"
				externalName := "externalName"
				resourceGetter := automock.NewInstanceListerByClusterServiceClass()
				resourceGetter.On("ListForClusterServiceClass", name, externalName, testCase.Namespace).Return(nil, expectedErr).Once()
				defer resourceGetter.AssertExpectations(t)

				parentObj := gqlschema.ClusterServiceClass{
					Name:         name,
					ExternalName: externalName,
				}

				resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, resourceGetter, nil, nil)

				_, err := resolver.ClusterServiceClassActivatedField(nil, &parentObj, testCase.Namespace)

				assert.Error(t, err)
				assert.True(t, gqlerror.IsInternal(err))
			})
		}

	})
}

func TestClusterServiceClassResolver_ClusterServiceClassPlansField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedSingleObj := gqlschema.ClusterServicePlan{
			Name: "Test",
		}
		expected := []gqlschema.ClusterServicePlan{
			expectedSingleObj,
			expectedSingleObj,
		}

		name := "name"
		resource := v1beta1.ClusterServicePlan{}
		resources := []*v1beta1.ClusterServicePlan{
			&resource,
			&resource,
		}
		resourceGetter := automock.NewClusterServicePlanLister()
		resourceGetter.On("ListForClusterServiceClass", name).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLClusterServicePlanConverter()
		converter.On("ToGQLs", resources).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}
		resolver := servicecatalog.NewClusterServiceClassResolver(nil, resourceGetter, nil, nil, nil)
		resolver.SetPlanConverter(converter)

		result, err := resolver.ClusterServiceClassPlansField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := automock.NewClusterServicePlanLister()
		resourceGetter.On("ListForClusterServiceClass", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}
		resolver := servicecatalog.NewClusterServiceClassResolver(nil, resourceGetter, nil, nil, nil)

		result, err := resolver.ClusterServiceClassPlansField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := automock.NewClusterServicePlanLister()
		resourceGetter.On("ListForClusterServiceClass", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}
		resolver := servicecatalog.NewClusterServiceClassResolver(nil, resourceGetter, nil, nil, nil)

		result, err := resolver.ClusterServiceClassPlansField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClassResolver_ClusterServiceClassClusterDocsTopicField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "name"
		resources := &v1alpha1.ClusterDocsTopic{
			ObjectMeta: v1.ObjectMeta{
				Name: name,
			},
		}
		expected := &gqlschema.ClusterDocsTopic{
			Name: name,
		}

		resourceGetter := new(mock.ClusterDocsTopicGetter)
		resourceGetter.On("Find", name).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := new(mock.GqlClusterDocsTopicConverter)
		converter.On("ToGQL", resources).Return(expected, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(mock.CmsRetriever)
		retriever.On("ClusterDocsTopic").Return(resourceGetter)
		retriever.On("ClusterDocsTopicConverter").Return(converter)

		parentObj := gqlschema.ClusterServiceClass{
			Name:         name,
			ExternalName: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, retriever, nil)

		result, err := resolver.ClusterServiceClassClusterDocsTopicField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"

		resourceGetter := new(mock.ClusterDocsTopicGetter)
		resourceGetter.On("Find", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := new(mock.GqlClusterDocsTopicConverter)
		converter.On("ToGQL", (*v1alpha1.ClusterDocsTopic)(nil)).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(mock.CmsRetriever)
		retriever.On("ClusterDocsTopic").Return(resourceGetter)
		retriever.On("ClusterDocsTopicConverter").Return(converter)

		parentObj := gqlschema.ClusterServiceClass{
			Name:         name,
			ExternalName: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, retriever, nil)

		result, err := resolver.ClusterServiceClassClusterDocsTopicField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"

		resourceGetter := new(mock.ClusterDocsTopicGetter)
		resourceGetter.On("Find", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(mock.CmsRetriever)
		retriever.On("ClusterDocsTopic").Return(resourceGetter)

		parentObj := gqlschema.ClusterServiceClass{
			Name:         name,
			ExternalName: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, retriever, nil)

		result, err := resolver.ClusterServiceClassClusterDocsTopicField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

//func TestClassResolver_ClusterServiceClassClusterAssetGroupField(t *testing.T) {
//	t.Run("Success", func(t *testing.T) {
//		name := "name"
//		resources := &rafterv1beta1.ClusterAssetGroup{
//			ObjectMeta: v1.ObjectMeta{
//				Name: name,
//			},
//		}
//		expected := &gqlschema.ClusterAssetGroup{
//			Name: name,
//		}
//
//		resourceGetter := new(mock.ClusterAssetGroupGetter)
//		resourceGetter.On("Find", name).Return(resources, nil).Once()
//		defer resourceGetter.AssertExpectations(t)
//
//		converter := new(mock.GqlClusterAssetGroupConverter)
//		converter.On("ToGQL", resources).Return(expected, nil).Once()
//		defer resourceGetter.AssertExpectations(t)
//
//		retriever := new(mock.RafterRetriever)
//		retriever.On("ClusterAssetGroup").Return(resourceGetter)
//		retriever.On("ClusterAssetGroupConverter").Return(converter)
//
//		parentObj := gqlschema.ClusterServiceClass{
//			Name:         name,
//			ExternalName: name,
//		}
//
//		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, retriever)
//
//		result, err := resolver.ClusterServiceClassClusterAssetGroupField(nil, &parentObj)
//
//		require.NoError(t, err)
//		assert.Equal(t, expected, result)
//	})
//
//	t.Run("NotFound", func(t *testing.T) {
//		name := "name"
//
//		resourceGetter := new(mock.ClusterAssetGroupGetter)
//		resourceGetter.On("Find", name).Return(nil, nil).Once()
//		defer resourceGetter.AssertExpectations(t)
//
//		converter := new(mock.GqlClusterAssetGroupConverter)
//		converter.On("ToGQL", (*rafterv1beta1.ClusterAssetGroup)(nil)).Return(nil, nil).Once()
//		defer resourceGetter.AssertExpectations(t)
//
//		retriever := new(mock.RafterRetriever)
//		retriever.On("ClusterAssetGroup").Return(resourceGetter)
//		retriever.On("ClusterAssetGroupConverter").Return(converter)
//
//		parentObj := gqlschema.ClusterServiceClass{
//			Name:         name,
//			ExternalName: name,
//		}
//
//		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, retriever)
//
//		result, err := resolver.ClusterServiceClassClusterAssetGroupField(nil, &parentObj)
//
//		require.NoError(t, err)
//		assert.Nil(t, result)
//	})
//
//	t.Run("Error", func(t *testing.T) {
//		expectedErr := errors.New("Test")
//		name := "name"
//
//		resourceGetter := new(mock.ClusterAssetGroupGetter)
//		resourceGetter.On("Find", name).Return(nil, expectedErr).Once()
//		defer resourceGetter.AssertExpectations(t)
//
//		retriever := new(mock.RafterRetriever)
//		retriever.On("ClusterAssetGroup").Return(resourceGetter)
//
//		parentObj := gqlschema.ClusterServiceClass{
//			Name:         name,
//			ExternalName: name,
//		}
//
//		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, retriever)
//
//		result, err := resolver.ClusterServiceClassClusterAssetGroupField(nil, &parentObj)
//
//		assert.Error(t, err)
//		assert.True(t, gqlerror.IsInternal(err))
//		assert.Nil(t, result)
//	})
//}

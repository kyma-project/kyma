package servicecatalog_test

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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

		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil, nil)
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

		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil, nil)

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

		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil, nil)

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

		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil, nil)
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
		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil, nil)
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
		resolver := servicecatalog.NewClusterServiceClassResolver(resourceGetter, nil, nil, nil, nil, nil)

		_, err := resolver.ClusterServiceClassesQuery(nil, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestClusterServiceClassResolver_ClusterServiceClassActivatedField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := true
		name := "name"
		externalName := "externalName"
		resources := []*v1beta1.ServiceInstance{{}, {}}
		resourceGetter := automock.NewInstanceListerByClusterServiceClass()
		resourceGetter.On("ListForClusterServiceClass", name, externalName).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name:         name,
			ExternalName: externalName,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, resourceGetter, nil, nil, nil)

		result, err := resolver.ClusterServiceClassActivatedField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		externalName := "externalName"
		resourceGetter := automock.NewInstanceListerByClusterServiceClass()
		resourceGetter.On("ListForClusterServiceClass", name, externalName).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := &gqlschema.ClusterServiceClass{
			Name:         name,
			ExternalName: externalName,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, resourceGetter, nil, nil, nil)

		result, err := resolver.ClusterServiceClassActivatedField(nil, parentObj)

		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		externalName := "externalName"
		resourceGetter := automock.NewInstanceListerByClusterServiceClass()
		resourceGetter.On("ListForClusterServiceClass", name, externalName).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name:         name,
			ExternalName: externalName,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, resourceGetter, nil, nil, nil)

		_, err := resolver.ClusterServiceClassActivatedField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
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
		resolver := servicecatalog.NewClusterServiceClassResolver(nil, resourceGetter, nil, nil, nil, nil)
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
		resolver := servicecatalog.NewClusterServiceClassResolver(nil, resourceGetter, nil, nil, nil, nil)

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
		resolver := servicecatalog.NewClusterServiceClassResolver(nil, resourceGetter, nil, nil, nil, nil)

		result, err := resolver.ClusterServiceClassPlansField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClusterServiceClassResolver_ClusterServiceClassContentField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "name"
		resource := &storage.Content{
			Raw: map[string]interface{}{
				"test": "data",
			},
		}
		expected := new(gqlschema.JSON)
		err := expected.UnmarshalGQL(resource.Raw)
		require.NoError(t, err)

		resourceGetter := new(automock.ContentGetter)
		resourceGetter.On("Find", "service-class", name).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, nil, nil, resourceGetter)

		result, err := resolver.ClusterServiceClassContentField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := new(automock.ContentGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, nil, nil, resourceGetter)

		result, err := resolver.ClusterServiceClassContentField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := new(automock.ContentGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, nil, nil, resourceGetter)

		result, err := resolver.ClusterServiceClassContentField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClusterServiceClassResolver_ClusterServiceClassApiSpecField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "name"
		resource := &storage.ApiSpec{
			Raw: map[string]interface{}{
				"test": "data",
			},
		}
		expected := new(gqlschema.JSON)
		err := expected.UnmarshalGQL(resource.Raw)
		require.NoError(t, err)

		resourceGetter := new(automock.ApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, nil, resourceGetter, nil)

		result, err := resolver.ClusterServiceClassApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := new(automock.ApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, nil, resourceGetter, nil)

		result, err := resolver.ClusterServiceClassApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := new(automock.ApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, nil, resourceGetter, nil)

		result, err := resolver.ClusterServiceClassApiSpecField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClusterServiceClassResolver_ClusterServiceClassAsyncApiSpecField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "name"
		resource := &storage.AsyncApiSpec{
			Raw: map[string]interface{}{
				"test": "data",
			},
		}
		expected := new(gqlschema.JSON)
		err := expected.UnmarshalGQL(resource.Raw)
		require.NoError(t, err)

		resourceGetter := new(automock.AsyncApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, resourceGetter, nil, nil)

		result, err := resolver.ClusterServiceClassAsyncApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := new(automock.AsyncApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, resourceGetter, nil, nil)

		result, err := resolver.ClusterServiceClassAsyncApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := new(automock.AsyncApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ClusterServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClusterServiceClassResolver(nil, nil, nil, resourceGetter, nil, nil)

		result, err := resolver.ClusterServiceClassAsyncApiSpecField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

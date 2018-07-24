package servicecatalog_test

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Create test suite to reduce boilerplate
func TestClassResolver_ServiceClassQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.ServiceClass{
			Name: "Test",
		}
		name := "name"
		resource := &v1beta1.ClusterServiceClass{}
		resourceGetter := automock.NewClassListGetter()
		resourceGetter.On("Find", name).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLClassConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewClassResolver(resourceGetter, nil, nil, nil, nil, nil)
		resolver.SetClassConverter(converter)

		result, err := resolver.ServiceClassQuery(nil, name)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := automock.NewClassListGetter()
		resourceGetter.On("Find", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := servicecatalog.NewClassResolver(resourceGetter, nil, nil, nil, nil, nil)

		result, err := resolver.ServiceClassQuery(nil, name)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"
		resource := &v1beta1.ClusterServiceClass{}
		resourceGetter := automock.NewClassListGetter()
		resourceGetter.On("Find", name).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := servicecatalog.NewClassResolver(resourceGetter, nil, nil, nil, nil, nil)

		result, err := resolver.ServiceClassQuery(nil, name)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestClassResolver_ServiceClassesQuery(t *testing.T) {
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
		expected := []gqlschema.ServiceClass{
			{
				Name: "Test",
			}, {
				Name: "Test",
			},
		}

		resourceGetter := automock.NewClassListGetter()
		resourceGetter.On("List", pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLClassConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewClassResolver(resourceGetter, nil, nil, nil, nil, nil)
		resolver.SetClassConverter(converter)

		result, err := resolver.ServiceClassesQuery(nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var resources []*v1beta1.ClusterServiceClass

		resourceGetter := automock.NewClassListGetter()
		resourceGetter.On("List", pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewClassResolver(resourceGetter, nil, nil, nil, nil, nil)
		var expected []gqlschema.ServiceClass

		result, err := resolver.ServiceClassesQuery(nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")

		var resources []*v1beta1.ClusterServiceClass

		resourceGetter := automock.NewClassListGetter()
		resourceGetter.On("List", pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewClassResolver(resourceGetter, nil, nil, nil, nil, nil)

		_, err := resolver.ServiceClassesQuery(nil, nil, nil)

		require.Error(t, err)
	})
}

func TestClassResolver_ServiceClassActivatedField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := true
		name := "name"
		externalName := "externalName"
		resources := []*v1beta1.ServiceInstance{{}, {}}
		resourceGetter := automock.NewClassInstanceLister()
		resourceGetter.On("ListForClass", name, externalName).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, resourceGetter, nil, nil, nil)

		result, err := resolver.ServiceClassActivatedField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		externalName := "externalName"
		resourceGetter := automock.NewClassInstanceLister()
		resourceGetter.On("ListForClass", name, externalName).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := &gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, resourceGetter, nil, nil, nil)

		result, err := resolver.ServiceClassActivatedField(nil, parentObj)

		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		externalName := "externalName"
		resourceGetter := automock.NewClassInstanceLister()
		resourceGetter.On("ListForClass", name, externalName).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, resourceGetter, nil, nil, nil)

		_, err := resolver.ServiceClassActivatedField(nil, &parentObj)

		assert.Error(t, err)
	})
}

func TestClassResolver_ServiceClassPlansField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedSingleObj := gqlschema.ServicePlan{
			Name: "Test",
		}
		expected := []gqlschema.ServicePlan{
			expectedSingleObj,
			expectedSingleObj,
		}

		name := "name"
		resource := v1beta1.ClusterServicePlan{}
		resources := []*v1beta1.ClusterServicePlan{
			&resource,
			&resource,
		}
		resourceGetter := automock.NewPlanLister()
		resourceGetter.On("ListForClass", name).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLPlanConverter()
		converter.On("ToGQLs", resources).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}
		resolver := servicecatalog.NewClassResolver(nil, resourceGetter, nil, nil, nil, nil)
		resolver.SetPlanConverter(converter)

		result, err := resolver.ServiceClassPlansField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := automock.NewPlanLister()
		resourceGetter.On("ListForClass", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}
		resolver := servicecatalog.NewClassResolver(nil, resourceGetter, nil, nil, nil, nil)

		result, err := resolver.ServiceClassPlansField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := automock.NewPlanLister()
		resourceGetter.On("ListForClass", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}
		resolver := servicecatalog.NewClassResolver(nil, resourceGetter, nil, nil, nil, nil)

		result, err := resolver.ServiceClassPlansField(nil, &parentObj)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestClassResolver_ServiceClassContentField(t *testing.T) {
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

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, nil, nil, nil, resourceGetter)

		result, err := resolver.ServiceClassContentField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := new(automock.ContentGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, nil, nil, nil, resourceGetter)

		result, err := resolver.ServiceClassContentField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := new(automock.ContentGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, nil, nil, nil, resourceGetter)

		result, err := resolver.ServiceClassContentField(nil, &parentObj)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestClassResolver_ServiceClassApiSpecField(t *testing.T) {
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

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, nil, nil, resourceGetter, nil)

		result, err := resolver.ServiceClassApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := new(automock.ApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, nil, nil, resourceGetter, nil)

		result, err := resolver.ServiceClassApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := new(automock.ApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, nil, nil, resourceGetter, nil)

		result, err := resolver.ServiceClassApiSpecField(nil, &parentObj)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestClassResolver_ServiceClassAsyncApiSpecField(t *testing.T) {
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

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, nil, resourceGetter, nil, nil)

		result, err := resolver.ServiceClassAsyncApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := new(automock.AsyncApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, nil, resourceGetter, nil, nil)

		result, err := resolver.ServiceClassAsyncApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := new(automock.AsyncApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewClassResolver(nil, nil, nil, resourceGetter, nil, nil)

		result, err := resolver.ServiceClassAsyncApiSpecField(nil, &parentObj)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

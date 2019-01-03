package servicecatalog_test

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/automock"
	contentMock "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/shared/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClassResolver_ServiceClassQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "name"
		env := "env"
		expected := &gqlschema.ServiceClass{
			Name: "Test",
		}
		resource := &v1beta1.ServiceClass{}
		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("Find", name, env).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLServiceClassConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)
		resolver.SetClassConverter(converter)

		result, err := resolver.ServiceClassQuery(nil, name, env)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		env := "env"
		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("Find", name, env).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)

		result, err := resolver.ServiceClassQuery(nil, name, env)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		name := "name"
		env := "env"
		expected := errors.New("Test")
		resource := &v1beta1.ServiceClass{}
		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("Find", name, env).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)

		result, err := resolver.ServiceClassQuery(nil, name, env)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClassResolver_ServiceClassesQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := "env"
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
		resourceGetter.On("List", env, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLServiceClassConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)
		resolver.SetClassConverter(converter)

		result, err := resolver.ServiceClassesQuery(nil, env, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		env := "env"
		var resources []*v1beta1.ServiceClass

		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("List", env, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)
		var expected []gqlschema.ServiceClass

		result, err := resolver.ServiceClassesQuery(nil, env, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		env := "env"
		expected := errors.New("Test")

		var resources []*v1beta1.ServiceClass

		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("List", env, pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewServiceClassResolver(resourceGetter, nil, nil, nil)

		_, err := resolver.ServiceClassesQuery(nil, env, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestClassResolver_ServiceClassActivatedField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := "env"
		name := "name"
		externalName := "externalName"
		resources := []*v1beta1.ServiceInstance{{}, {}}
		resourceGetter := automock.NewInstanceListerByServiceClass()
		resourceGetter.On("ListForServiceClass", name, externalName, env).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
			Environment:  env,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, resourceGetter, nil)

		result, err := resolver.ServiceClassActivatedField(nil, &parentObj)

		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		env := "env"
		name := "name"
		externalName := "externalName"
		resourceGetter := automock.NewInstanceListerByServiceClass()
		resourceGetter.On("ListForServiceClass", name, externalName, env).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := &gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
			Environment:  env,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, resourceGetter, nil)

		result, err := resolver.ServiceClassActivatedField(nil, parentObj)

		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		env := "env"
		expectedErr := errors.New("Test")
		name := "name"
		externalName := "externalName"
		resourceGetter := automock.NewInstanceListerByServiceClass()
		resourceGetter.On("ListForServiceClass", name, externalName, env).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:         name,
			ExternalName: externalName,
			Environment:  env,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, resourceGetter, nil)

		_, err := resolver.ServiceClassActivatedField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestClassResolver_ServiceClassPlansField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := "env"
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
		resourceGetter.On("ListForServiceClass", name, env).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLServicePlanConverter()
		converter.On("ToGQLs", resources).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:        name,
			Environment: env,
		}
		resolver := servicecatalog.NewServiceClassResolver(nil, resourceGetter, nil, nil)
		resolver.SetPlanConverter(converter)

		result, err := resolver.ServiceClassPlansField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		env := "env"
		name := "name"
		resourceGetter := automock.NewServicePlanLister()
		resourceGetter.On("ListForServiceClass", name, env).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:        name,
			Environment: env,
		}
		resolver := servicecatalog.NewServiceClassResolver(nil, resourceGetter, nil, nil)

		result, err := resolver.ServiceClassPlansField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		env := "env"
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := automock.NewServicePlanLister()
		resourceGetter.On("ListForServiceClass", name, env).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceClass{
			Name:        name,
			Environment: env,
		}
		resolver := servicecatalog.NewServiceClassResolver(nil, resourceGetter, nil, nil)

		result, err := resolver.ServiceClassPlansField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
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

		resourceGetter := new(contentMock.ContentGetter)
		resourceGetter.On("Find", "service-class", name).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("Content").Return(resourceGetter)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassContentField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := new(contentMock.ContentGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("Content").Return(resourceGetter)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassContentField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := new(contentMock.ContentGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("Content").Return(resourceGetter)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassContentField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
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

		resourceGetter := new(contentMock.ApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("ApiSpec").Return(resourceGetter)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := new(contentMock.ApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("ApiSpec").Return(resourceGetter)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := new(contentMock.ApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("ApiSpec").Return(resourceGetter)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassApiSpecField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
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

		resourceGetter := new(contentMock.AsyncApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("AsyncApiSpec").Return(resourceGetter)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassAsyncApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := new(contentMock.AsyncApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("AsyncApiSpec").Return(resourceGetter)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassAsyncApiSpecField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := new(contentMock.AsyncApiSpecGetter)
		resourceGetter.On("Find", "service-class", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("AsyncApiSpec").Return(resourceGetter)

		parentObj := gqlschema.ServiceClass{
			Name: name,
		}

		resolver := servicecatalog.NewServiceClassResolver(nil, nil, nil, retriever)

		result, err := resolver.ServiceClassAsyncApiSpecField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

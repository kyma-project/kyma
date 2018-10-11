package servicecatalog_test

import (
	"encoding/json"
	"testing"

	api "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestServiceBindingResolver_CreateServiceBindingMutation(t *testing.T) {
	t.Run("Success empty name", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		binding := fixServiceBindingToRedis()
		binding.Namespace = ""
		binding.Name = ""
		svc.On("Create", "production", binding).
			Return(fixServiceBindingToRedis(), nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		result, err := resolver.CreateServiceBindingMutation(nil, nil, "redis", "production", nil)

		require.NoError(t, err)
		assert.Equal(t, fixCreateServiceBindingOutput(), result)
	})

	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		params := gqlschema.JSON(map[string]interface{}{
			"key": "value",
		})
		binding := fixServiceBindingToRedis()
		binding.Namespace = ""
		binding.Name = "generated-name"
		byteArray, err := json.Marshal(params)
		require.NoError(t, err)
		binding.Spec.Parameters = &runtime.RawExtension{
			Raw: byteArray,
		}
		svc.On("Create", "production", binding).
			Return(fixServiceBindingToRedis(), nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		result, err := resolver.CreateServiceBindingMutation(nil, ptr("generated-name"), "redis", "production", &params)

		require.NoError(t, err)
		assert.Equal(t, fixCreateServiceBindingOutput(), result)
	})

	t.Run("Already exists", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		svc.On("Create", mock.Anything, mock.Anything).Return(nil, apiErrors.NewAlreadyExists(schema.GroupResource{}, "test")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		_, err := resolver.CreateServiceBindingMutation(nil, ptr("redis-binding"), "redis", "production", nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsAlreadyExists(err))
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		svc.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("nope")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		_, err := resolver.CreateServiceBindingMutation(nil, ptr("redis-binding"), "redis", "production", nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceBindingResolver_DeleteServiceBindingMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		svc.On("Delete", "production", "redis-binding").Return(nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		result, err := resolver.DeleteServiceBindingMutation(nil, "redis-binding", "production")

		require.NoError(t, err)
		assert.Equal(t, &gqlschema.DeleteServiceBindingOutput{
			Name:        "redis-binding",
			Environment: "production",
		}, result)
	})

	t.Run("Not found", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		svc.On("Delete", "production", "redis-binding").Return(apiErrors.NewNotFound(schema.GroupResource{}, "test")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		_, err := resolver.DeleteServiceBindingMutation(nil, "redis-binding", "production")

		require.Error(t, err)
		assert.True(t, gqlerror.IsNotFound(err))
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		svc.On("Delete", "production", "redis-binding").Return(errors.New("ta")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		_, err := resolver.DeleteServiceBindingMutation(nil, "redis-binding", "production")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceBindingResolver_ServiceBindingQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		svc.On("Find", "production", "redis-binding").
			Return(fixServiceBindingToRedis(), nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		result, err := resolver.ServiceBindingQuery(nil, "redis-binding", "production")

		require.NoError(t, err)
		assert.Equal(t, fixServiceBindingGQLToRedis(), result)
	})

	t.Run("Not found", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		svc.On("Find", "production", "redis-binding").
			Return(nil, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		result, err := resolver.ServiceBindingQuery(nil, "redis-binding", "production")

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		svc.On("Find", "production", "redis-binding").
			Return(nil, errors.New("nope")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		_, err := resolver.ServiceBindingQuery(nil, "redis-binding", "production")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceBindingResolver_ServiceBindingsToInstanceQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		svc.On("ListForServiceInstance", "production", "redis").
			Return([]*api.ServiceBinding{
				fixServiceBindingToRedis(),
				fixServiceBindingToRedis(),
			}, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		result, err := resolver.ServiceBindingsToInstanceQuery(nil, "redis", "production")

		require.NoError(t, err)
		assert.Equal(t, gqlschema.ServiceBindings{
			Items: []gqlschema.ServiceBinding{
				*fixServiceBindingGQLToRedis(),
				*fixServiceBindingGQLToRedis(),
			},
			Stats: gqlschema.ServiceBindingsStats{
				Pending: 2,
			},
		}, result)
	})

	t.Run("Not found", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		svc.On("ListForServiceInstance", "production", "redis").Return([]*api.ServiceBinding{}, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		result, err := resolver.ServiceBindingsToInstanceQuery(nil, "redis", "production")

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingOperations()
		svc.On("ListForServiceInstance", "production", "redis").Return(nil, errors.New("yhm")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingResolver(svc)

		_, err := resolver.ServiceBindingsToInstanceQuery(nil, "redis", "production")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func fixServiceBindingGQLToRedis() *gqlschema.ServiceBinding {
	return &gqlschema.ServiceBinding{
		Name:                "redis-binding",
		ServiceInstanceName: "redis",
		Environment:         "production",
		Status: gqlschema.ServiceBindingStatus{
			Type: gqlschema.ServiceBindingStatusTypePending,
		},
	}
}

func fixCreateServiceBindingOutput() *gqlschema.CreateServiceBindingOutput {
	return &gqlschema.CreateServiceBindingOutput{
		Environment:         "production",
		ServiceInstanceName: "redis",
		Name:                "redis-binding",
	}
}

func ptr(s string) *string {
	return &s
}

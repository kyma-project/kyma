package servicecatalogaddons_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	api "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestServiceBindingUsageResolver_CreateServiceBindingUsageMutation(t *testing.T) {
	const namespace = "test-ns"
	t.Run("Success with empty name", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		bindingUsage := fixServiceBindingUsageResource()
		bindingUsage.Namespace = ""
		bindingUsage.Name = ""
		svc.On("Create", "test-ns", bindingUsage).Return(fixServiceBindingUsageResource(), nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		input := fixCreateServiceBindingUsageInput()
		input.Name = nil
		result, err := resolver.CreateServiceBindingUsageMutation(nil, namespace, input)

		require.NoError(t, err)
		assert.Equal(t, fixServiceBindingUsage(), result)
	})

	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		bindingUsage := fixServiceBindingUsageResource()
		bindingUsage.Namespace = ""
		svc.On("Create", "test-ns", bindingUsage).Return(fixServiceBindingUsageResource(), nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		input := fixCreateServiceBindingUsageInput()
		result, err := resolver.CreateServiceBindingUsageMutation(nil, namespace, input)

		require.NoError(t, err)
		assert.Equal(t, fixServiceBindingUsage(), result)
	})

	t.Run("Already exists", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Create", mock.Anything, mock.Anything).Return(nil, apiErrors.NewAlreadyExists(schema.GroupResource{}, "test")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)
		binding := fixCreateServiceBindingUsageInput()

		_, err := resolver.CreateServiceBindingUsageMutation(nil, namespace, binding)

		require.Error(t, err)
		assert.True(t, gqlerror.IsAlreadyExists(err))
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("trololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		_, err := resolver.CreateServiceBindingUsageMutation(nil, namespace, fixCreateServiceBindingUsageInput())

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceBindingUsageResolver_DeleteServiceBindingUsageMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Delete", "test", "test").Return(nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		result, err := resolver.DeleteServiceBindingUsageMutation(nil, "test", "test")

		require.NoError(t, err)
		assert.Equal(t, &gqlschema.DeleteServiceBindingUsageOutput{
			Name:      "test",
			Namespace: "test",
		}, result)
	})

	t.Run("Not exists", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Delete", "test", "test").Return(apiErrors.NewNotFound(schema.GroupResource{}, "test")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		_, err := resolver.DeleteServiceBindingUsageMutation(nil, "test", "test")

		require.Error(t, err)
		assert.True(t, gqlerror.IsNotFound(err))
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Delete", "test", "test").Return(errors.New("trololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		_, err := resolver.DeleteServiceBindingUsageMutation(nil, "test", "test")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceBindingUsageResolver_ServiceBindingUsageQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Find", "test", "test").Return(fixServiceBindingUsageResource(), nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		result, err := resolver.ServiceBindingUsageQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Equal(t, fixServiceBindingUsage(), result)
	})

	t.Run("Not found", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Find", "test", "test").Return(nil, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		result, err := resolver.ServiceBindingUsageQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Find", "test", "test").Return(nil, errors.New("trolololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		_, err := resolver.ServiceBindingUsageQuery(nil, "test", "test")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceBindingUsageResolver_ServiceBindingUsagesOfInstanceQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		usages := []*api.ServiceBindingUsage{
			fixServiceBindingUsageResource(),
			fixServiceBindingUsageResource(),
		}
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("ListForServiceInstance", "test", "test").Return(usages, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		result, err := resolver.ServiceBindingUsagesOfInstanceQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Equal(t, []gqlschema.ServiceBindingUsage{
			*fixServiceBindingUsage(),
			*fixServiceBindingUsage(),
		}, result)
	})

	t.Run("Not found", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("ListForServiceInstance", "test", "test").Return([]*api.ServiceBindingUsage{}, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		result, err := resolver.ServiceBindingUsagesOfInstanceQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("ListForServiceInstance", "test", "test").Return(nil, errors.New("trolololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc)

		_, err := resolver.ServiceBindingUsagesOfInstanceQuery(nil, "test", "test")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func fixServiceBindingUsage() *gqlschema.ServiceBindingUsage {
	return &gqlschema.ServiceBindingUsage{
		Name:      "sbu-name",
		Namespace: "test-ns",
		UsedBy: gqlschema.LocalObjectReference{
			Kind: "Deployment",
			Name: "sample-deployment",
		},
		ServiceBindingName: "binding-name",
		Status: gqlschema.ServiceBindingUsageStatus{
			Type: gqlschema.ServiceBindingUsageStatusTypePending,
		},
	}
}

func fixCreateServiceBindingUsageInput() *gqlschema.CreateServiceBindingUsageInput {
	name := "sbu-name"
	return &gqlschema.CreateServiceBindingUsageInput{
		Name: &name,
		ServiceBindingRef: gqlschema.ServiceBindingRefInput{
			Name: "binding-name",
		},
		UsedBy: gqlschema.LocalObjectReferenceInput{
			Kind: "Deployment",
			Name: "sample-deployment",
		},
	}
}

func fixServiceBindingUsageResource() *api.ServiceBindingUsage {
	return &api.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingUsage",
			APIVersion: "servicecatalog.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sbu-name",
			Namespace: "test-ns",
		},
		Spec: api.ServiceBindingUsageSpec{
			ServiceBindingRef: api.LocalReferenceByName{
				Name: "binding-name",
			},
			UsedBy: api.LocalReferenceByKindAndName{
				Kind: "Deployment",
				Name: "sample-deployment",
			},
		},
	}
}

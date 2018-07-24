package servicecatalog_test

import (
	"errors"
	"fmt"
	"testing"

	api "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stesting "k8s.io/client-go/testing"
)

func TestServiceBindingUsageResolver_CreateServiceBindingUsageMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		bindingUsage := fixServiceBindingUsageResource()
		bindingUsage.Namespace = ""
		svc.On("Create", "test-ns", bindingUsage).Return(fixServiceBindingUsageResource(), nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)

		result, err := resolver.CreateServiceBindingUsageMutation(nil, fixCreateServiceBindingUsageInput())

		require.NoError(t, err)
		assert.Equal(t, fixServiceBindingUsage(), result)
	})

	t.Run("Wrong kind", func(t *testing.T) {
		resolver := servicecatalog.NewServiceBindingUsageResolver(nil)

		input := fixCreateServiceBindingUsageInput()
		input.UsedBy.Kind = "nope"
		_, err := resolver.CreateServiceBindingUsageMutation(nil, input)

		require.Error(t, err)
	})

	t.Run("Already exists", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Create", mock.Anything, mock.Anything).Return(nil, apiErrors.NewAlreadyExists(schema.GroupResource{}, "test")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)
		binding := fixCreateServiceBindingUsageInput()

		_, err := resolver.CreateServiceBindingUsageMutation(nil, binding)

		require.Error(t, err)
		assert.Equal(t, fmt.Sprintf("ServiceBindingUsage %s already exists", binding.Name), err.Error())
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("trololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)

		_, err := resolver.CreateServiceBindingUsageMutation(nil, fixCreateServiceBindingUsageInput())

		require.Error(t, err)
	})
}

func TestServiceBindingUsageResolver_DeleteServiceBindingUsageMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Delete", "test", "test").Return(nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)

		result, err := resolver.DeleteServiceBindingUsageMutation(nil, "test", "test")

		require.NoError(t, err)
		assert.Equal(t, &gqlschema.DeleteServiceBindingUsageOutput{
			Name:        "test",
			Environment: "test",
		}, result)
	})

	t.Run("Not exists", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Delete", "test", "test").Return(apiErrors.NewNotFound(schema.GroupResource{}, "test")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)

		_, err := resolver.DeleteServiceBindingUsageMutation(nil, "test", "test")

		require.Error(t, err)
		assert.Equal(t, "ServiceBindingUsage test not found", err.Error())
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Delete", "test", "test").Return(errors.New("trololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)

		_, err := resolver.DeleteServiceBindingUsageMutation(nil, "test", "test")

		require.Error(t, err)
	})
}

func TestServiceBindingUsageResolver_ServiceBindingUsageQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Find", "test", "test").Return(fixServiceBindingUsageResource(), nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)

		result, err := resolver.ServiceBindingUsageQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Equal(t, fixServiceBindingUsage(), result)
	})

	t.Run("Not found", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Find", "test", "test").Return(nil, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)

		result, err := resolver.ServiceBindingUsageQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("Find", "test", "test").Return(nil, errors.New("trolololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)

		_, err := resolver.ServiceBindingUsageQuery(nil, "test", "test")

		require.Error(t, err)
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
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)

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
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)

		result, err := resolver.ServiceBindingUsagesOfInstanceQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		svc.On("ListForServiceInstance", "test", "test").Return(nil, errors.New("trolololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBindingUsageResolver(svc)

		_, err := resolver.ServiceBindingUsagesOfInstanceQuery(nil, "test", "test")

		require.Error(t, err)
	})
}

func fixServiceBindingUsage() *gqlschema.ServiceBindingUsage {
	return &gqlschema.ServiceBindingUsage{
		Name:        "bu-name",
		Environment: "test-ns",
		UsedBy: gqlschema.LocalObjectReference{
			Kind: gqlschema.BindingUsageReferenceTypeDeployment,
			Name: "sample-deployment",
		},
		ServiceBindingName: "binding-name",
		Status: gqlschema.ServiceBindingUsageStatus{
			Type: gqlschema.ServiceBindingUsageStatusTypePending,
		},
	}
}

func fixCreateServiceBindingUsageInput() *gqlschema.CreateServiceBindingUsageInput {
	return &gqlschema.CreateServiceBindingUsageInput{
		Name:        "bu-name",
		Environment: "test-ns",
		ServiceBindingRef: gqlschema.ServiceBindingRefInput{
			Name: "binding-name",
		},
		UsedBy: gqlschema.LocalObjectReferenceInput{
			Kind: gqlschema.BindingUsageReferenceTypeDeployment,
			Name: "sample-deployment",
		},
	}
}

func fixServiceBindingUsageResource() *api.ServiceBindingUsage {
	return &api.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingUsage",
			APIVersion: "servicecatalog.kyma.cx/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bu-name",
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

func failingReactor(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
	return true, nil, errors.New("custom error")
}

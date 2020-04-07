package servicecatalogaddons_test

import (
	"context"
	"errors"
	"testing"
	"time"

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
		converter := automock.NewServiceBindingUsageConverter()

		unchangedBindingUsage := fixServiceBindingUsageResource()
		bindingUsage := fixServiceBindingUsageResource()
		bindingUsage.Namespace = ""
		bindingUsage.Name = ""
		svc.On("Create", "test-ns", bindingUsage).Return(unchangedBindingUsage, nil).Once()
		defer svc.AssertExpectations(t)

		gqlBindingUsage := fixCreateServiceBindingUsageInput()
		gqlBindingUsage.Name = nil
		converter.On("InputToK8s", gqlBindingUsage).Return(bindingUsage, nil).Once()
		converter.On("ToGQL", unchangedBindingUsage).Return(fixServiceBindingUsage(), nil).Once()
		defer svc.AssertExpectations(t)

		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		result, err := resolver.CreateServiceBindingUsageMutation(nil, namespace, gqlBindingUsage)

		require.NoError(t, err)
		assert.Equal(t, fixServiceBindingUsage(), result)
	})

	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		unchangedBindingUsage := fixServiceBindingUsageResource()
		bindingUsage := fixServiceBindingUsageResource()
		bindingUsage.Namespace = ""
		svc.On("Create", "test-ns", bindingUsage).Return(unchangedBindingUsage, nil).Once()
		defer svc.AssertExpectations(t)

		gqlBindingUsage := fixCreateServiceBindingUsageInput()
		converter.On("InputToK8s", gqlBindingUsage).Return(bindingUsage, nil).Once()
		converter.On("ToGQL", unchangedBindingUsage).Return(fixServiceBindingUsage(), nil).Once()
		defer svc.AssertExpectations(t)

		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		result, err := resolver.CreateServiceBindingUsageMutation(nil, namespace, gqlBindingUsage)

		require.NoError(t, err)
		assert.Equal(t, fixServiceBindingUsage(), result)
	})

	t.Run("Already exists", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		gqlBindingUsage := fixCreateServiceBindingUsageInput()
		bindingUsage := fixServiceBindingUsageResource()
		converter.On("InputToK8s", gqlBindingUsage).Return(bindingUsage, nil).Once()
		defer svc.AssertExpectations(t)

		svc.On("Create", mock.Anything, mock.Anything).Return(nil, apiErrors.NewAlreadyExists(schema.GroupResource{}, "test")).Once()
		defer svc.AssertExpectations(t)

		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		_, err := resolver.CreateServiceBindingUsageMutation(nil, namespace, gqlBindingUsage)

		require.Error(t, err)
		assert.True(t, gqlerror.IsAlreadyExists(err))
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		gqlBindingUsage := fixCreateServiceBindingUsageInput()
		bindingUsage := fixServiceBindingUsageResource()
		converter.On("InputToK8s", gqlBindingUsage).Return(bindingUsage, nil).Once()
		defer svc.AssertExpectations(t)

		svc.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("trololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		_, err := resolver.CreateServiceBindingUsageMutation(nil, namespace, gqlBindingUsage)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceBindingUsageResolver_DeleteServiceBindingUsageMutation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		svc.On("Delete", "test", "test").Return(nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		result, err := resolver.DeleteServiceBindingUsageMutation(nil, "test", "test")

		require.NoError(t, err)
		assert.Equal(t, &gqlschema.DeleteServiceBindingUsageOutput{
			Name:      "test",
			Namespace: "test",
		}, result)
	})

	t.Run("Not exists", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		svc.On("Delete", "test", "test").Return(apiErrors.NewNotFound(schema.GroupResource{}, "test")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		_, err := resolver.DeleteServiceBindingUsageMutation(nil, "test", "test")

		require.Error(t, err)
		assert.True(t, gqlerror.IsNotFound(err))
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		svc.On("Delete", "test", "test").Return(errors.New("trololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		_, err := resolver.DeleteServiceBindingUsageMutation(nil, "test", "test")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceBindingUsageResolver_DeleteServiceBindingUsagesMutation(t *testing.T) {
	usages := []string{"test1", "test2"}

	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		for _, usage := range usages {
			svc.On("Delete", "test", usage).Return(nil).Once()
		}
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		result, err := resolver.DeleteServiceBindingUsagesMutation(nil, usages, "test")

		require.NoError(t, err)
		assert.Equal(t, []*gqlschema.DeleteServiceBindingUsageOutput{
			{
				Name:      "test1",
				Namespace: "test",
			},
			{
				Name:      "test2",
				Namespace: "test",
			},
		}, result)
	})

	t.Run("Not exists", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		svc.On("Delete", "test", "test1").Return(nil).Once()
		svc.On("Delete", "test", "test2").Return(apiErrors.NewNotFound(schema.GroupResource{}, "test")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		_, err := resolver.DeleteServiceBindingUsagesMutation(nil, usages, "test")

		require.Error(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		svc.On("Delete", "test", "test1").Return(nil).Once()
		svc.On("Delete", "test", "test2").Return(errors.New("trololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		_, err := resolver.DeleteServiceBindingUsagesMutation(nil, usages, "test")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceBindingUsageResolver_ServiceBindingUsageQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		bindingUsage := fixServiceBindingUsageResource()
		gqlBindingUsage := fixServiceBindingUsage()
		converter.On("ToGQL", bindingUsage).Return(gqlBindingUsage, nil).Once()
		defer svc.AssertExpectations(t)

		svc.On("Find", "test", "test").Return(bindingUsage, nil).Once()
		defer svc.AssertExpectations(t)

		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		result, err := resolver.ServiceBindingUsageQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Equal(t, fixServiceBindingUsage(), result)
	})

	t.Run("Not found", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		nilUsage := (*api.ServiceBindingUsage)(nil)

		converter.On("ToGQL", nilUsage).Return(nil, nil).Once()
		defer svc.AssertExpectations(t)

		svc.On("Find", "test", "test").Return(nil, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		result, err := resolver.ServiceBindingUsageQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		svc.On("Find", "test", "test").Return(nil, errors.New("trolololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

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
		gqlUsages := []gqlschema.ServiceBindingUsage{
			*fixServiceBindingUsage(),
			*fixServiceBindingUsage(),
		}

		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		converter.On("ToGQLs", usages).Return(gqlUsages, nil).Once()
		defer svc.AssertExpectations(t)

		svc.On("ListForServiceInstance", "test", "test").Return(usages, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		result, err := resolver.ServiceBindingUsagesOfInstanceQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Equal(t, gqlUsages, result)
	})

	t.Run("Not found", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		usages := []*api.ServiceBindingUsage{}
		gqlUsages := []gqlschema.ServiceBindingUsage{}

		converter.On("ToGQLs", usages).Return(gqlUsages, nil).Once()
		defer svc.AssertExpectations(t)

		svc.On("ListForServiceInstance", "test", "test").Return(usages, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		result, err := resolver.ServiceBindingUsagesOfInstanceQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()

		svc.On("ListForServiceInstance", "test", "test").Return(nil, errors.New("trolololo")).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		_, err := resolver.ServiceBindingUsagesOfInstanceQuery(nil, "test", "test")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceBindingUsageResolver_ServiceBindingUsageEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		_, err := resolver.ServiceBindingUsageEventSubscription(ctx, "test", nil, nil)

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewServiceBindingUsageOperations()
		converter := automock.NewServiceBindingUsageConverter()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := servicecatalogaddons.NewServiceBindingUsageResolver(svc, converter)

		channel, err := resolver.ServiceBindingUsageEventSubscription(ctx, "test", nil, nil)
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
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

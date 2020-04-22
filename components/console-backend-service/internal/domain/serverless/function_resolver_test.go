package serverless

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/automock"
	shared "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestFunctionResolver_FunctionQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		labels := map[string]string{"foo": "bar"}
		function := fixFunction("1", "a", "1", "content", "dependencies", labels)
		gqlFunction := fixGQLFunction("1", "a", "1", "content", "dependencies", labels)

		svc := automock.NewFunctionService()
		svc.On("Find", "a", "1").Return(function, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLFunctionConverter()
		converter.On("ToGQL", function).Return(&gqlFunction, nil)
		defer converter.AssertExpectations(t)

		resolver := newFunctionResolver(svc, converter, nil, nil)

		result, err := resolver.FunctionQuery(nil, "1", "a")
		require.NoError(t, err)
		assert.Equal(t, &gqlFunction, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var resource *v1alpha1.Function
		var expected *gqlschema.Function

		svc := automock.NewFunctionService()
		svc.On("Find", "a", "1").Return(resource, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLFunctionConverter()
		converter.On("ToGQL", resource).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := newFunctionResolver(svc, converter, nil, nil)

		result, err := resolver.FunctionQuery(nil, "1", "a")
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Error")
		var resource *v1alpha1.Function

		svc := automock.NewFunctionService()
		svc.On("Find", "a", "1").Return(resource, expected).Once()
		defer svc.AssertExpectations(t)

		resolver := newFunctionResolver(svc, nil, nil, nil)

		_, err := resolver.FunctionQuery(nil, "1", "a")
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestFunctionResolver_FunctionsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		labels := map[string]string{"foo": "bar"}
		function1 := fixFunction("1", "a", "1", "content", "dependencies", labels)
		function2 := fixFunction("2", "a", "1", "content", "dependencies", labels)
		gqlFunction1 := fixGQLFunction("1", "a", "1", "content", "dependencies", labels)
		gqlFunction2 := fixGQLFunction("2", "a", "1", "content", "dependencies", labels)
		functions := []*v1alpha1.Function{function1, function2}
		expected := []gqlschema.Function{gqlFunction1, gqlFunction2}

		svc := automock.NewFunctionService()
		svc.On("List", "a").Return(functions, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLFunctionConverter()
		converter.On("ToGQLs", functions).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := newFunctionResolver(svc, converter, nil, nil)

		result, err := resolver.FunctionsQuery(nil, "a")
		require.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var resources []*v1alpha1.Function
		var expected []gqlschema.Function

		svc := automock.NewFunctionService()
		svc.On("List", "a").Return(resources, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLFunctionConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := newFunctionResolver(svc, converter, nil, nil)

		result, err := resolver.FunctionsQuery(nil, "a")
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Error")
		var resources []*v1alpha1.Function

		svc := automock.NewFunctionService()
		svc.On("List", "a").Return(resources, expected).Once()
		defer svc.AssertExpectations(t)

		resolver := newFunctionResolver(svc, nil, nil, nil)

		_, err := resolver.FunctionsQuery(nil, "a")
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestFunctionResolver_CreateFunction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		labels := map[string]string{"foo": "bar"}
		function := fixFunction("1", "a", "1", "content", "dependencies", labels)
		gqlFunction := fixGQLFunction("1", "a", "1", "content", "dependencies", labels)
		mutationInput := fixGQLMutationInput("content", "dependencies", labels)

		svc := automock.NewFunctionService()
		svc.On("Create", function).Return(function, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLFunctionConverter()
		converter.On("ToFunction", "1", "a", mutationInput).Return(function, nil)
		converter.On("ToGQL", function).Return(&gqlFunction, nil)
		defer converter.AssertExpectations(t)

		resolver := newFunctionResolver(svc, converter, nil, nil)

		result, err := resolver.CreateFunction(nil, "1", "a", mutationInput)
		require.NoError(t, err)
		assert.Equal(t, &gqlFunction, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Error")
		labels := map[string]string{"foo": "bar"}
		function := fixFunction("1", "a", "1", "content", "dependencies", labels)
		mutationInput := fixGQLMutationInput("content", "dependencies", labels)

		svc := automock.NewFunctionService()
		svc.On("Create", function).Return(function, expected).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLFunctionConverter()
		converter.On("ToFunction", "1", "a", mutationInput).Return(function, nil)
		defer converter.AssertExpectations(t)

		resolver := newFunctionResolver(svc, converter, nil, nil)

		result, err := resolver.CreateFunction(nil, "1", "a", mutationInput)
		require.Error(t, err)
		require.Nil(t, result)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestFunctionResolver_UpdateFunction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		labels := map[string]string{"foo": "bar"}
		function := fixFunction("1", "a", "1", "content", "dependencies", labels)
		gqlFunction := fixGQLFunction("1", "a", "1", "content", "dependencies", labels)
		mutationInput := fixGQLMutationInput("content", "dependencies", labels)

		svc := automock.NewFunctionService()
		svc.On("Update", function).Return(function, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLFunctionConverter()
		converter.On("ToFunction", "1", "a", mutationInput).Return(function, nil)
		converter.On("ToGQL", function).Return(&gqlFunction, nil)
		defer converter.AssertExpectations(t)

		resolver := newFunctionResolver(svc, converter, nil, nil)

		result, err := resolver.UpdateFunction(nil, "1", "a", mutationInput)
		require.NoError(t, err)
		assert.Equal(t, &gqlFunction, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Error")
		labels := map[string]string{"foo": "bar"}
		function := fixFunction("1", "a", "1", "content", "dependencies", labels)
		mutationInput := fixGQLMutationInput("content", "dependencies", labels)

		svc := automock.NewFunctionService()
		svc.On("Update", function).Return(function, expected).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLFunctionConverter()
		converter.On("ToFunction", "1", "a", mutationInput).Return(function, nil)
		defer converter.AssertExpectations(t)

		resolver := newFunctionResolver(svc, converter, nil, nil)

		result, err := resolver.UpdateFunction(nil, "1", "a", mutationInput)
		require.Error(t, err)
		require.Nil(t, result)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestFunctionResolver_DeleteFunction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		usageKind := "lambda"
		mutationInput := fixGQLMetadataInput("1", "a")
		mutation := fixGQLMetadata("1", "a")

		svc := automock.NewFunctionService()
		svc.On("Delete", mutationInput).Return(nil).Once()
		defer svc.AssertExpectations(t)

		resourceLister := new(shared.ServiceBindingUsageLister)
		resourceLister.On("DeleteAllByUsageKind", mutationInput.Namespace, usageKind, mutationInput.Name).Return(nil).Once()
		defer resourceLister.AssertExpectations(t)

		retriever := new(shared.ServiceCatalogAddonsRetriever)
		retriever.On("ServiceBindingUsage").Return(resourceLister)

		resolver := newFunctionResolver(svc, nil, &Config{UsageKind: usageKind}, retriever)

		result, err := resolver.DeleteFunction(nil, mutationInput)
		require.NoError(t, err)
		assert.Equal(t, &mutation, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Error")
		mutationInput := fixGQLMetadataInput("1", "a")

		svc := automock.NewFunctionService()
		svc.On("Delete", mutationInput).Return(expected).Once()
		defer svc.AssertExpectations(t)

		resolver := newFunctionResolver(svc, nil, nil, nil)

		result, err := resolver.DeleteFunction(nil, mutationInput)
		require.Error(t, err)
		require.Nil(t, result)
	})
}

func TestFunctionResolver_DeleteManyFunction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		usageKind := "lambda"
		mutationInput1 := fixGQLMetadataInput("1", "a")
		mutationInput2 := fixGQLMetadataInput("2", "a")
		resources := []gqlschema.FunctionMetadataInput{mutationInput1, mutationInput2}

		mutation1 := fixGQLMetadata("1", "a")
		mutation2 := fixGQLMetadata("2", "a")
		mutations := []gqlschema.FunctionMetadata{mutation1, mutation2}

		svc := automock.NewFunctionService()
		svc.On("Delete", mutationInput1).Return(nil).Once()
		svc.On("Delete", mutationInput2).Return(nil).Once()
		defer svc.AssertExpectations(t)

		resourceLister := new(shared.ServiceBindingUsageLister)
		resourceLister.On("DeleteAllByUsageKind", mutationInput1.Namespace, usageKind, mutationInput1.Name).Return(nil).Once()
		resourceLister.On("DeleteAllByUsageKind", mutationInput2.Namespace, usageKind, mutationInput2.Name).Return(nil).Once()
		defer resourceLister.AssertExpectations(t)

		retriever := new(shared.ServiceCatalogAddonsRetriever)
		retriever.On("ServiceBindingUsage").Return(resourceLister)

		resolver := newFunctionResolver(svc, nil, &Config{UsageKind: usageKind}, retriever)

		result, err := resolver.DeleteManyFunctions(nil, resources)
		require.NoError(t, err)
		assert.Equal(t, mutations, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Error")
		mutationInput1 := fixGQLMetadataInput("1", "a")
		mutationInput2 := fixGQLMetadataInput("2", "a")
		resources := []gqlschema.FunctionMetadataInput{mutationInput1, mutationInput2}

		svc := automock.NewFunctionService()
		svc.On("Delete", mutationInput1).Return(expected).Once()
		defer svc.AssertExpectations(t)

		resolver := newFunctionResolver(svc, nil, nil, nil)

		result, err := resolver.DeleteManyFunctions(nil, resources)
		require.Error(t, err)
		require.Equal(t, []gqlschema.FunctionMetadata{}, result)
	})
}

func TestFunctionResolver_FunctionEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewFunctionService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := newFunctionResolver(svc, nil, nil, nil)

		_, err := resolver.FunctionEventSubscription(ctx, "", nil)

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewFunctionService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := newFunctionResolver(svc, nil, nil, nil)

		channel, err := resolver.FunctionEventSubscription(ctx, "", nil)
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}

func fixFakeFunctionService(t *testing.T, objects ...runtime.Object) *functionService {
	serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme, objects...)
	require.NoError(t, err)

	service := newFunctionService(serviceFactory)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

	return service
}

func fixGQLMetadataInput(name, namespace string) gqlschema.FunctionMetadataInput {
	return gqlschema.FunctionMetadataInput{
		Name:      name,
		Namespace: namespace,
	}
}

func fixGQLMetadata(name, namespace string) gqlschema.FunctionMetadata {
	return gqlschema.FunctionMetadata{
		Name:      name,
		Namespace: namespace,
	}
}

func fixGQLMutationInput(source, dependencies string, labels map[string]string) gqlschema.FunctionMutationInput {
	return gqlschema.FunctionMutationInput{
		Labels:       labels,
		Source:       source,
		Dependencies: dependencies,
	}
}

func fixGQLFunction(name, namespace, uid, source, dependencies string, labels map[string]string) gqlschema.Function {
	return gqlschema.Function{
		Name:         name,
		Namespace:    namespace,
		UID:          uid,
		Labels:       labels,
		Source:       source,
		Dependencies: dependencies,
		Env: []gqlschema.FunctionEnv{
			{
				Name:  "foo",
				Value: "bar",
			},
		},
		Replicas:  gqlschema.FunctionReplicas{},
		Resources: gqlschema.FunctionResources{},
		Status: gqlschema.FunctionStatus{
			Phase: gqlschema.FunctionPhaseTypeInitializing,
		},
	}
}

func fixFunction(name, namespace, uid, source, dependencies string, labels map[string]string) *v1alpha1.Function {
	return &v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "serverless.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
			UID:       types.UID(uid),
		},
		Spec: v1alpha1.FunctionSpec{
			Source: source,
			Deps:   dependencies,
			Env: []v1.EnvVar{
				{
					Name:  "foo",
					Value: "bar",
				},
			},
		},
		Status: v1alpha1.FunctionStatus{},
	}
}

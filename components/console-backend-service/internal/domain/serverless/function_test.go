package serverless

import (
	"context"
	"errors"
	"testing"

	. "github.com/golang/mock/gomock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/function"
	mockserverless "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/mocks"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/testing/prop"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var testName = prop.String(16)
var testNamespace = prop.String(16)
var testLabels = gqlschema.Labels{prop.String(5): prop.String(5), prop.String(5): prop.String(5)}
var testSize = prop.OneOfString("S", "M", "L")
var testRuntime = prop.OneOfString("nodejs6", "nodejs8")
var testContent = prop.String(50)
var testDependencies = prop.String(50)

func TestResolver_FunctionsQuery(t *testing.T) {
	functionA := fixFunction("a", testName, testNamespace, testSize, testRuntime, testContent, testDependencies, testLabels)
	functionB := fixFunction("b", testName, testNamespace, testSize, testRuntime, testContent, testDependencies, testLabels)
	functionC := fixFunction("c", testName, testNamespace, testSize, testRuntime, testContent, testDependencies, testLabels)

	expected := []gqlschema.Function{*function.ToGQL(functionA), *function.ToGQL(functionB), *function.ToGQL(functionC)}

	t.Run("Returns functions always in the same order", func(t *testing.T) {
		ctrl := NewController(t)
		defer ctrl.Finish()
		service := mockserverless.NewMockFunctionService(ctrl)
		resolver := resolver{functionService: service}

		service.EXPECT().
			List(Eq(testNamespace)).
			Return([]*v1alpha1.Function{functionA, functionB, functionC}, nil)

		query, err := resolver.FunctionsQuery(context.TODO(), testNamespace)
		require.NoError(t, err)
		assert.Equal(t, expected, query)

		service.EXPECT().
			List(Eq(testNamespace)).
			Return([]*v1alpha1.Function{functionC, functionA, functionB}, nil)

		query, err = resolver.FunctionsQuery(context.TODO(), testNamespace)
		require.NoError(t, err)
		assert.Equal(t, expected, query)
	})
}

func TestResolver_FunctionQuery(t *testing.T) {
	fixedFunction := fixFunction("a", testName, testNamespace, testSize, testRuntime, testContent, testDependencies, testLabels)
	expected := function.ToGQL(fixedFunction)

	t.Run("Success", func(t *testing.T) {
		ctrl := NewController(t)
		defer ctrl.Finish()
		service := mockserverless.NewMockFunctionService(ctrl)
		resolver := resolver{functionService: service}

		service.EXPECT().
			Find(Eq(testName), Eq(testNamespace)).
			Return(fixedFunction, nil)

		query, err := resolver.FunctionQuery(context.TODO(), testName, testNamespace)
		require.NoError(t, err)
		assert.Equal(t, expected, query)
	})

	t.Run("Not found", func(t *testing.T) {
		ctrl := NewController(t)
		defer ctrl.Finish()
		service := mockserverless.NewMockFunctionService(ctrl)
		resolver := resolver{functionService: service}

		service.EXPECT().
			Find(Eq(testName), Eq(testNamespace)).
			Return((*v1alpha1.Function)(nil), nil)

		query, err := resolver.FunctionQuery(context.TODO(), testName, testNamespace)
		require.NoError(t, err)
		assert.Nil(t, query)
	})

	t.Run("Error", func(t *testing.T) {
		ctrl := NewController(t)
		defer ctrl.Finish()
		service := mockserverless.NewMockFunctionService(ctrl)
		resolver := resolver{functionService: service}

		service.EXPECT().
			Find(Eq(testName), Eq(testNamespace)).
			Return((*v1alpha1.Function)(nil), errors.New("internal error"))

		query, err := resolver.FunctionQuery(context.TODO(), testName, testNamespace)
		require.Error(t, err)
		assert.Nil(t, query)
	})
}

func TestResolver_CreateFunction(t *testing.T) {
	fixedFunction := fixFunction("a", testName, testNamespace, testSize, testRuntime, testContent, testDependencies, testLabels)
	expected := function.ToGQL(fixedFunction)

	t.Run("Success", func(t *testing.T) {
		ctrl := NewController(t)
		defer ctrl.Finish()
		service := mockserverless.NewMockFunctionService(ctrl)
		resolver := resolver{functionService: service}

		service.EXPECT().
			Create(Eq(testName), Eq(testNamespace), Eq(testLabels), Eq(testSize), Eq(testRuntime)).
			Return(fixedFunction, nil)

		mutation, err := resolver.CreateFunction(context.TODO(), testName, testNamespace, testLabels, testSize, testRuntime)

		require.NoError(t, err)
		assert.Equal(t, expected, mutation)
	})

	t.Run("Error", func(t *testing.T) {
		ctrl := NewController(t)
		defer ctrl.Finish()
		service := mockserverless.NewMockFunctionService(ctrl)
		resolver := resolver{functionService: service}

		service.EXPECT().
			Create(Eq(testName), Eq(testNamespace), Eq(testLabels), Eq(testSize), Eq(testRuntime)).
			Return((*v1alpha1.Function)(nil), errors.New("already exists"))

		mutation, err := resolver.CreateFunction(context.TODO(), testName, testNamespace, testLabels, testSize, testRuntime)
		require.Error(t, err)
		assert.Nil(t, mutation)
	})
}

func TestResolver_UpdateFunction(t *testing.T) {
	newLabels := gqlschema.Labels{prop.String(5): prop.String(5), prop.String(5): prop.String(5)}
	newSize := prop.OneOfString("S", "M", "L")
	newRuntime := prop.OneOfString("nodejs6", "nodejs8")
	newContent := prop.String(50)
	newDependencies := prop.String(50)
	params := gqlschema.FunctionUpdateInput{Labels: newLabels, Size: newSize, Runtime: newRuntime, Content: newContent, Dependencies: newDependencies}

	fixedFunction := fixFunction("a", testName, testNamespace, newSize, newRuntime, newContent, newDependencies, newLabels)
	expected := function.ToGQL(fixedFunction)

	t.Run("Success", func(t *testing.T) {
		ctrl := NewController(t)
		defer ctrl.Finish()
		service := mockserverless.NewMockFunctionService(ctrl)
		resolver := resolver{functionService: service}

		service.EXPECT().
			Update(Eq(testName), Eq(testNamespace), Eq(params)).
			Return(fixedFunction, nil)

		mutation, err := resolver.UpdateFunction(context.TODO(), testName, testNamespace, params)
		require.NoError(t, err)
		assert.Equal(t, expected, mutation)
	})

	t.Run("Error", func(t *testing.T) {
		ctrl := NewController(t)
		defer ctrl.Finish()
		service := mockserverless.NewMockFunctionService(ctrl)
		resolver := resolver{functionService: service}

		service.EXPECT().
			Update(Eq(testName), Eq(testNamespace), Eq(params)).
			Return((*v1alpha1.Function)(nil), errors.New("not found"))

		mutation, err := resolver.UpdateFunction(context.TODO(), testName, testNamespace, params)
		require.Error(t, err)
		assert.Nil(t, mutation)
	})
}

func TestResolver_DeleteFunction(t *testing.T) {
	expected := gqlschema.FunctionMutationOutput{
		Name:      testName,
		Namespace: testNamespace,
	}

	t.Run("Success", func(t *testing.T) {
		ctrl := NewController(t)
		defer ctrl.Finish()
		service := mockserverless.NewMockFunctionService(ctrl)
		resolver := resolver{functionService: service}

		service.EXPECT().
			Delete(Eq(testName), Eq(testNamespace)).
			Return(nil)

		mutation, err := resolver.DeleteFunction(context.TODO(), testName, testNamespace)
		require.NoError(t, err)
		assert.Equal(t, &expected, mutation)
	})

	t.Run("Error", func(t *testing.T) {
		ctrl := NewController(t)
		defer ctrl.Finish()
		service := mockserverless.NewMockFunctionService(ctrl)
		resolver := resolver{functionService: service}

		service.EXPECT().
			Delete(Eq(testName), Eq(testNamespace)).
			Return(errors.New("not found"))

		mutation, err := resolver.DeleteFunction(context.TODO(), testName, testNamespace)
		require.Error(t, err)
		assert.Nil(t, mutation)
	})
}

func fixFunction(uid, name, namespace, size, runtime, content, dependencies string, labels gqlschema.Labels) *v1alpha1.Function {
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
			Size:     size,
			Runtime:  runtime,
			Function: content,
			Deps:     dependencies,
		},
	}
}

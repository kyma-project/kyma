package serverless

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestFunctionService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		labels := map[string]string{"foo": "bar"}
		function1 := fixFunction("1", "a", "1", "content", "dependencies", labels)
		function2 := fixFunction("2", "a", "2", "content", "dependencies", labels)
		function3 := fixFunction("3", "b", "3", "content", "dependencies", labels)

		service := fixFakeFunctionService(t, function1, function2, function3)

		result, err := service.Find("a", "1")
		require.NoError(t, err)
		assert.Equal(t, function1, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		service := fixFakeFunctionService(t)

		result, err := service.Find("a", "1")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestFunctionService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		labels := map[string]string{"foo": "bar"}
		function1 := fixFunction("1", "a", "1", "content", "dependencies", labels)
		function2 := fixFunction("2", "a", "2", "content", "dependencies", labels)
		function3 := fixFunction("3", "b", "3", "content", "dependencies", labels)
		expected := []*v1alpha1.Function{
			function1,
			function2,
		}

		service := fixFakeFunctionService(t, function1, function2, function3)

		result, err := service.List("a")
		require.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		service := fixFakeFunctionService(t)

		result, err := service.List("a")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestFunctionService_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		labels := map[string]string{"foo": "bar"}
		function1 := fixFunction("1", "a", "1", "content", "dependencies", labels)

		service := fixFakeFunctionService(t)

		result, err := service.Create(function1)
		require.NoError(t, err)
		assert.Equal(t, function1, result)
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		labels := map[string]string{"foo": "bar"}
		function1 := fixFunction("1", "a", "1", "content", "dependencies", labels)

		service := fixFakeFunctionService(t, function1)

		result, err := service.Create(function1)
		assert.True(t, apiErrors.IsAlreadyExists(err))
		assert.Nil(t, result)
	})
}

func TestFunctionService_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		oldLabels := map[string]string{"foo": "bar"}
		newLabels := map[string]string{"bar": "foo"}
		function1 := fixFunction("1", "a", "1", "content", "dependencies", oldLabels)
		function2 := fixFunction("1", "a", "1", "content", "dependencies", newLabels)

		service := fixFakeFunctionService(t, function1)

		result, err := service.Update(function2)
		require.NoError(t, err)
		assert.Equal(t, function2, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		labels := map[string]string{"foo": "bar"}
		function1 := fixFunction("1", "a", "1", "content", "dependencies", labels)

		service := fixFakeFunctionService(t)

		result, err := service.Update(function1)
		assert.True(t, apiErrors.IsNotFound(err))
		assert.Nil(t, result)
	})
}

func TestFunctionService_Delete(t *testing.T) {
	labels := map[string]string{"foo": "bar"}
	function1 := fixFunction("1", "a", "1", "content", "dependencies", labels)
	function2 := fixFunction("2", "a", "2", "content", "dependencies", labels)
	function3 := fixFunction("3", "b", "3", "content", "dependencies", labels)

	for testName, testData := range map[string]struct {
		function gqlschema.FunctionMetadataInput
		error    bool
	}{
		"Success": {
			function: gqlschema.FunctionMetadataInput{Name: "1", Namespace: "a"},
			error:    false,
		},
		"Without namespace": {
			function: gqlschema.FunctionMetadataInput{Name: "a2"},
			error:    true,
		},
		"Without name": {
			function: gqlschema.FunctionMetadataInput{Namespace: "a"},
			error:    true,
		},
		"Empty": {
			function: gqlschema.FunctionMetadataInput{},
			error:    true,
		},
	} {
		t.Run(testName, func(t *testing.T) {
			service := fixFakeFunctionService(t, function1, function2, function3)

			err := service.Delete(testData.function)
			if testData.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFunctionService_DeleteMany(t *testing.T) {
	labels := map[string]string{"foo": "bar"}
	function1 := fixFunction("1", "a", "1", "content", "dependencies", labels)
	function2 := fixFunction("2", "a", "2", "content", "dependencies", labels)
	function3 := fixFunction("3", "b", "3", "content", "dependencies", labels)

	for testName, testData := range map[string]struct {
		functions []gqlschema.FunctionMetadataInput
		error     bool
	}{
		"Success": {
			functions: []gqlschema.FunctionMetadataInput{
				{Name: "1", Namespace: "a"},
				{Name: "2", Namespace: "a"},
			},
			error: false,
		},
		"Error": {
			functions: []gqlschema.FunctionMetadataInput{
				{Name: "1", Namespace: "a"},
				{Name: "2"},
			},
			error: true,
		},
		"Empty": {
			functions: []gqlschema.FunctionMetadataInput{},
			error:     false,
		},
	} {
		t.Run(testName, func(t *testing.T) {
			service := fixFakeFunctionService(t, function1, function2, function3)

			err := service.DeleteMany(testData.functions)
			if testData.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFunctionService_SubscribeAndUnsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		service := fixFakeFunctionService(t)

		functionListener := newFunctionListener(nil, nil, nil)
		service.Subscribe(functionListener)
		service.Unsubscribe(functionListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		service := fixFakeFunctionService(t)

		functionListener := newFunctionListener(nil, nil, nil)
		service.Subscribe(functionListener)
		service.Subscribe(functionListener)

		service.Unsubscribe(functionListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		service := fixFakeFunctionService(t)

		functionListenerA := newFunctionListener(nil, nil, nil)
		functionListenerB := newFunctionListener(nil, nil, nil)

		service.Subscribe(functionListenerA)
		service.Subscribe(functionListenerB)

		service.Unsubscribe(functionListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixFakeFunctionService(t)

		service.Subscribe(nil)
		service.Unsubscribe(nil)
	})
}

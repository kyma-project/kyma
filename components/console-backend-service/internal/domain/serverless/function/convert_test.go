package function_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/function"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvert_FunctionToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedName := "expectedName"
		expectedNamespace := "expectedNamespace"
		expectedRuntime := "expectedRuntime"
		expectedSize := "expectedSize"
		expectedLabels := gqlschema.Labels{"test": "label"}
		expectedStatus := gqlschema.FunctionStatusTypeUpdating

		in := v1alpha1.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name:      expectedName,
				Labels:    expectedLabels,
				Namespace: expectedNamespace,
			},
			Spec: v1alpha1.FunctionSpec{
				Runtime: expectedRuntime,
				Size:    expectedSize,
			},
			Status: v1alpha1.FunctionStatus{
				Condition: v1alpha1.FunctionCondition(expectedStatus),
			},
		}

		result := function.ToGQL(&in)

		assert.Equal(t, expectedName, result.Name)
		assert.Equal(t, expectedNamespace, result.Namespace)
		assert.Equal(t, expectedRuntime, result.Runtime)
		assert.Equal(t, expectedSize, result.Size)
		assert.Equal(t, expectedLabels, result.Labels)
		assert.Equal(t, expectedStatus, result.Status)
	})

	t.Run("Empty", func(t *testing.T) {
		in := v1alpha1.Function{}
		expected := gqlschema.Function{
			Status: gqlschema.FunctionStatusTypeError}
		result := function.ToGQL(&in)
		assert.Equal(t, &expected, result)
	})
}

func TestConvert_FunctionsToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedName := "expectedName"
		expectedNamespace := "expectedNamespace"
		expectedLabels := gqlschema.Labels{"test": "label"}

		expectedName2 := "expectedName2"
		expectedNamespace2 := "expectedNamespace2"
		expectedLabels2 := gqlschema.Labels{"test": "label"}

		in := []*v1alpha1.Function{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      expectedName,
					Labels:    expectedLabels,
					Namespace: expectedNamespace,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      expectedName2,
					Labels:    expectedLabels2,
					Namespace: expectedNamespace2,
				},
			},
		}

		result := function.ToGQLs(in)

		assert.Len(t, result, 2)
		assert.Equal(t, expectedName, result[0].Name)
		assert.Equal(t, expectedLabels, result[0].Labels)
		assert.Equal(t, expectedNamespace, result[0].Namespace)
		assert.Equal(t, expectedName2, result[1].Name)
		assert.Equal(t, expectedLabels2, result[1].Labels)
		assert.Equal(t, expectedNamespace2, result[1].Namespace)

	})

	t.Run("With nil", func(t *testing.T) {
		expectedName := "expectedName"
		expectedNamespace := "expectedNamespace"
		expectedLabels := gqlschema.Labels{"test": "label"}

		in := []*v1alpha1.Function{
			nil,
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      expectedName,
					Labels:    expectedLabels,
					Namespace: expectedNamespace,
				},
			},
			nil,
		}

		result := function.ToGQLs(in)

		assert.Len(t, result, 1)
		assert.Equal(t, expectedName, result[0].Name)
		assert.Equal(t, expectedLabels, result[0].Labels)
		assert.Equal(t, expectedNamespace, result[0].Namespace)
	})

	t.Run("Empty", func(t *testing.T) {
		var in []*v1alpha1.Function

		result := function.ToGQLs(in)

		assert.Empty(t, result)
	})
}

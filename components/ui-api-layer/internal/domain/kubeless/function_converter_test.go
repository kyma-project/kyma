package kubeless

import (
	"testing"
	"time"

	"github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFunctionConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		var zeroTimeStamp time.Time
		function := fixFunction()

		expected := &gqlschema.Function{
			Name:              "test",
			Labels:            gqlschema.JSON{"test": "ok", "ok": "test"},
			CreationTimestamp: zeroTimeStamp,
			Trigger:           "nope",
			Environment:       "env",
		}

		converter := functionConverter{}
		result := converter.ToGQL(function)

		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := functionConverter{}
		result := converter.ToGQL(&v1beta1.Function{})

		require.NotNil(t, result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := functionConverter{}
		result := converter.ToGQL(nil)

		require.Nil(t, result)
	})
}

func TestFunctionConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		functions := []*v1beta1.Function{
			fixFunction(),
			fixFunction(),
		}

		converter := functionConverter{}
		result := converter.ToGQLs(functions)

		assert.Len(t, result, 2)
		assert.Equal(t, "test", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var functions []*v1beta1.Function

		converter := functionConverter{}
		result := converter.ToGQLs(functions)

		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		functions := []*v1beta1.Function{
			nil,
			fixFunction(),
			nil,
		}

		converter := functionConverter{}
		result := converter.ToGQLs(functions)

		assert.Len(t, result, 1)
		assert.Equal(t, "test", result[0].Name)
	})
}

func fixFunction() *v1beta1.Function {
	var mockTimeStamp v1.Time

	return &v1beta1.Function{
		ObjectMeta: v1.ObjectMeta{
			Name:              "test",
			CreationTimestamp: mockTimeStamp,
			Labels: map[string]string{
				"test": "ok",
				"ok":   "test",
			},
			Namespace: "env",
		},
		Spec: v1beta1.FunctionSpec{
			Type: "nope",
		},
	}
}

package serverless

import (
	"context"
	. "github.com/golang/mock/gomock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/function"
	mock_serverless "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/mocks"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/testing/prop"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"

)

func TestResolver_FunctionsQuery(t *testing.T) {
	testNamespace := prop.String(16)
	functionA := fixFunction("a")
	functionB := fixFunction("b")
	functionC := fixFunction("c")
	expected := []gqlschema.Function{*function.ToGQL(functionA), *function.ToGQL(functionB), *function.ToGQL(functionC)}

	t.Run("Returns functions always in the same order", func(t *testing.T) {
		ctrl := NewController(t)
		defer ctrl.Finish()
		service := mock_serverless.NewMockFunctionService(ctrl)
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

func fixFunction(uid string) *v1alpha1.Function {
	return &v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "serverless.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      prop.String(4),
			Namespace: prop.String(4),
			Labels:    prop.Labels(2),
			UID:       types.UID(uid),
		},
		Spec: v1alpha1.FunctionSpec{
			Size:    prop.OneOfString("S", "M", "L"),
			Runtime: prop.OneOfString("nodejs6", "nodejs8"),
		},
	}
}

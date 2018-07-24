package kubeless_test

import (
	"testing"

	"github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/kubeless"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/kubeless/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFunctionResolver_FunctionsQuery(t *testing.T) {
	environment := "test"
	pagingParams := pager.PagingParams{}

	t.Run("Success", func(t *testing.T) {
		function := &v1beta1.Function{
			ObjectMeta: v1.ObjectMeta{
				Name: "test",
			},
		}
		functions := []*v1beta1.Function{
			function, function,
		}

		expected := gqlschema.Function{
			Name:   "test",
			Labels: gqlschema.JSON{},
		}

		svc := automock.NewFunctionLister()
		svc.On("List", environment, pagingParams).Return(functions, nil).Once()

		resolver := kubeless.NewFunctionResolver(svc)

		result, err := resolver.FunctionsQuery(nil, environment, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, []gqlschema.Function{expected, expected}, result)

	})

	t.Run("Not found", func(t *testing.T) {
		var functions []*v1beta1.Function
		var expected []gqlschema.Function

		svc := automock.NewFunctionLister()
		svc.On("List", environment, pagingParams).Return(functions, nil).Once()

		resolver := kubeless.NewFunctionResolver(svc)

		result, err := resolver.FunctionsQuery(nil, environment, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewFunctionLister()
		svc.On("List", environment, pagingParams).Return(nil, errors.New("test")).Once()

		resolver := kubeless.NewFunctionResolver(svc)

		_, err := resolver.FunctionsQuery(nil, environment, nil, nil)
		require.Error(t, err)
	})
}

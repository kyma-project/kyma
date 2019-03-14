package kubeless_test

import (
	"testing"

	"github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/kubeless"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/kubeless/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFunctionResolver_FunctionsQuery(t *testing.T) {
	namespace := "test"
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
			Labels: gqlschema.Labels(nil),
		}

		svc := automock.NewFunctionLister()
		svc.On("List", namespace, pagingParams).Return(functions, nil).Once()

		resolver, err := kubeless.NewFunctionResolver(svc)
		require.NoError(t, err)

		result, err := resolver.FunctionsQuery(nil, namespace, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, []gqlschema.Function{expected, expected}, result)

	})

	t.Run("Not found", func(t *testing.T) {
		var functions []*v1beta1.Function
		var expected []gqlschema.Function

		svc := automock.NewFunctionLister()
		svc.On("List", namespace, pagingParams).Return(functions, nil).Once()

		resolver, err := kubeless.NewFunctionResolver(svc)
		require.NoError(t, err)

		result, err := resolver.FunctionsQuery(nil, namespace, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewFunctionLister()
		svc.On("List", namespace, pagingParams).Return(nil, errors.New("test")).Once()

		resolver, err := kubeless.NewFunctionResolver(svc)
		require.NoError(t, err)

		_, err = resolver.FunctionsQuery(nil, namespace, nil, nil)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

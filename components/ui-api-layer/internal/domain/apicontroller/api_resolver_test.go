package apicontroller

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/v1alpha2"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/apicontroller/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApiResolver_APIsQuery(t *testing.T) {
	environment := "test-1"

	t.Run("Should return a list of APIs", func(t *testing.T) {
		apis := []*v1alpha2.Api{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-1",
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-2",
				},
			},
		}

		expected := []gqlschema.API{
			{
				Name: apis[0].Name,
			},
			{
				Name: apis[1].Name,
			},
		}

		var empty *string = nil

		service := automock.NewApiLister()
		service.On("List", environment, empty, empty).Return(apis, nil).Once()

		resolver := newApiResolver(service)

		result, err := resolver.APIsQuery(nil, environment, nil, nil)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Should return an error", func(t *testing.T) {
		var empty *string = nil

		service := automock.NewApiLister()
		service.On("List", environment, empty, empty).Return(nil, errors.New("test")).Once()

		resolver := newApiResolver(service)

		_, err := resolver.APIsQuery(nil, environment, nil, nil)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.Equal(t, "cannot query APIs", err.Error())
	})
}

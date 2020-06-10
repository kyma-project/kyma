package k8s_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/types"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceResolver_CreateResourceMutation(t *testing.T) {
	const (
		apiVersion = "v1"
		kind       = "Pod"
		name       = "test-name"
		namespace  = "test-namespace"
	)

	var (
		resourceJSON = gqlschema.JSON{
			"kind":       kind,
			"apiVersion": apiVersion,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		}
		resource = types.Resource{
			APIVersion: apiVersion,
			Name:       name,
			Namespace:  namespace,
			Kind:       kind,
		}
	)

	t.Run("Success", func(t *testing.T) {
		body, err := json.Marshal(resource)
		require.NoError(t, err)
		createdResource := resource
		createdResource.Body = body

		svc := automock.NewResourceSvc()
		svc.On("Create", namespace, resource).Return(&createdResource, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLResourceConverter()
		converter.On("GQLJSONToResource", resourceJSON).Return(resource, nil).Once()
		converter.On("BodyToGQLJSON", createdResource.Body).Return(resourceJSON, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewResourceResolver(svc)
		resolver.SetResourceConverter(converter)

		result, err := resolver.CreateResourceMutation(nil, namespace, resourceJSON)

		require.NoError(t, err)
		assert.Equal(t, resourceJSON, result)
	})

	t.Run("ErrorConvertingToResource", func(t *testing.T) {
		expected := errors.New("fix")

		svc := automock.NewResourceSvc()

		converter := automock.NewGQLResourceConverter()
		converter.On("GQLJSONToResource", resourceJSON).Return(types.Resource{}, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewResourceResolver(svc)
		resolver.SetResourceConverter(converter)

		result, err := resolver.CreateResourceMutation(nil, namespace, resourceJSON)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorCreatingResource", func(t *testing.T) {
		expected := errors.New("fix")

		body, err := json.Marshal(resource)
		require.NoError(t, err)
		createdResource := resource
		createdResource.Body = body

		svc := automock.NewResourceSvc()
		svc.On("Create", namespace, resource).Return(nil, expected).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLResourceConverter()
		converter.On("GQLJSONToResource", resourceJSON).Return(resource, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewResourceResolver(svc)
		resolver.SetResourceConverter(converter)

		result, err := resolver.CreateResourceMutation(nil, namespace, resourceJSON)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConvertingBodyToGQLJSON", func(t *testing.T) {
		expected := errors.New("fix")

		body, err := json.Marshal(resource)
		require.NoError(t, err)
		createdResource := resource
		createdResource.Body = body

		svc := automock.NewResourceSvc()
		svc.On("Create", namespace, resource).Return(&createdResource, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLResourceConverter()
		converter.On("GQLJSONToResource", resourceJSON).Return(resource, nil).Once()
		converter.On("BodyToGQLJSON", createdResource.Body).Return(gqlschema.JSON{}, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewResourceResolver(svc)
		resolver.SetResourceConverter(converter)

		result, err := resolver.CreateResourceMutation(nil, namespace, resourceJSON)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

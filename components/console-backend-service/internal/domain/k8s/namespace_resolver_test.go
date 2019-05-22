package k8s_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	appAutomock "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestNamespaceResolver_NamespacesQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "name"
		labels := map[string]string{
			"env": "true",
		}
		resource := fixNamespace(name, labels)
		resources := []*v1.Namespace{resource, resource}
		expectedResult := gqlschema.Namespace{
			Name:   name,
			Labels: labels,
		}

		expected := []gqlschema.Namespace{
			expectedResult, expectedResult,
		}
		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("List").Return(resources, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewNamespaceConverter()
		converter.On("ToGQLs", resources).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever)
		resolver.SetNamespaceConverter(converter)

		result, err := resolver.NamespacesQuery(nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		resources := []*v1.Namespace{}
		expected := []gqlschema.Namespace(nil)
		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("List").Return(resources, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewNamespaceConverter()
		converter.On("ToGQLs", resources).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever)
		resolver.SetNamespaceConverter(converter)

		result, err := resolver.NamespacesQuery(nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("ErrorListing", func(t *testing.T) {
		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("List").Return(nil, errors.New("Error")).Once()
		defer svc.AssertExpectations(t)
		resolver := k8s.NewNamespaceResolver(svc, appRetriever)

		result, err := resolver.NamespacesQuery(nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})

	t.Run("ErrorConverting", func(t *testing.T) {
		resources := []*v1.Namespace{}
		expected := errors.New("Test")
		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("List").Return(resources, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewNamespaceConverter()
		converter.On("ToGQLs", resources).Return(nil, expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewNamespaceResolver(svc, appRetriever)
		resolver.SetNamespaceConverter(converter)

		result, err := resolver.NamespacesQuery(nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestNamespaceResolver_CreateNamespace(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "exampleName"
		labels := gqlschema.Labels{
			"test": "true",
		}
		expectedLabels := gqlschema.Labels{
			"test": "true",
			"env":  "true",
		}
		resource := fixNamespace(name, expectedLabels)
		expected := gqlschema.NamespaceCreationOutput{
			Name:   name,
			Labels: expectedLabels,
		}

		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Create", name, expectedLabels).Return(resource, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := k8s.NewNamespaceResolver(svc, appRetriever)
		result, err := resolver.CreateNamespace(nil, name, &labels)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		name := "exampleName"
		labels := gqlschema.Labels{}
		expectedLabels := gqlschema.Labels{
			"env": "true",
		}

		svc := automock.NewNamespaceSvc()
		appRetriever := new(appAutomock.ApplicationRetriever)
		svc.On("Create", name, expectedLabels).Return(nil, errors.New("Error")).Once()
		defer svc.AssertExpectations(t)
		resolver := k8s.NewNamespaceResolver(svc, appRetriever)

		result, err := resolver.CreateNamespace(nil, name, &labels)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.NotNil(t, result)
	})
}

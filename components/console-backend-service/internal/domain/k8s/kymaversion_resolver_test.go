package k8s_test

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKymaVersionResolver_KymaVersionQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := "version"

		deployment := fixDeploymentWithImage()
		svc := automock.NewKymaVersionSvc()
		svc.On("FindDeployment").Return(deployment, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGqlKymaVersionConverter()
		converter.On("ToKymaVersion", deployment.Spec.Template.Spec.Containers[0].Image).Return(expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewKymaVersionResolver(svc)
		resolver.SetKymaVersionConverter(converter)

		result, err := resolver.KymaVersionQuery(nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
	//t.Run("NotFound", func(t *testing.T) {
	//	svc := automock.NewKymaVersionSvc()
	//	svc.On("FindDeployment").Return(nil, nil).Once()
	//	defer svc.AssertExpectations(t)
	//
	//	resolver := k8s.NewKymaVersionResolver(svc)
	//	result, err := resolver.KymaVersionQuery(nil)
	//
	//	require.Error(t, err)
	//	assert.Equal(t, "", result)
	//})
}
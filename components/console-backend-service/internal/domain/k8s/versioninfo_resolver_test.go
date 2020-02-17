package k8s_test

import (
	"fmt"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestVersionInfoResolver_VersionInfoQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		version := "version"
		expected := gqlschema.VersionInfo{
			KymaVersion: version,
		}

		deployment := fixDeploymentWithImage()
		svc := automock.NewDeploymentLister()
		svc.On("Find", "kyma-installer", "kyma-installer").Return(deployment, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGqlVersionInfoConverter()
		converter.On("ToGQL", deployment).Return(expected).Once()
		defer converter.AssertExpectations(t)

		resolver := k8s.NewVersionInfoResolver(svc)
		resolver.SetVersionInfoConverter(converter)

		result, err := resolver.VersionInfoQuery(nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
	t.Run("NotFound", func(t *testing.T) {
		svc := automock.NewDeploymentLister()
		svc.On("Find", "kyma-installer", "kyma-installer").Return(nil, fmt.Errorf("error")).Once()
		defer svc.AssertExpectations(t)

		resolver := k8s.NewVersionInfoResolver(svc)
		result, err := resolver.VersionInfoQuery(nil)

		require.Error(t, err)
		assert.Equal(t, gqlschema.VersionInfo{}, result)
	})
}

func fixDeploymentWithImage() *apps.Deployment {
	return &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kyma-installer",
			Namespace: "kyma-installer",
		},
		Spec: apps.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "image",
						},
					},
				},
			},
		},
	}
}

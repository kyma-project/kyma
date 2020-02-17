package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"testing"

	"github.com/stretchr/testify/assert"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestVersionInfoConverter_ToGQL(t *testing.T) {
	t.Run("Non eu.gcr.io version", func(t *testing.T) {
		image := "test-repo/test-image"
		expected := gqlschema.VersionInfo{
			KymaVersion: image,
		}

		converter := &versionInfoConverter{}
		deployment := fixDeploymentWithImage(image)
		result := converter.ToGQL(deployment)

		assert.Equal(t, expected, result)
	})

	t.Run("Scemantic version", func(t *testing.T) {
		image := "eu.gcr.io/test/1.2.3"
		expected := gqlschema.VersionInfo{
			KymaVersion: "1.2.3",
		}

		converter := &versionInfoConverter{}
		deployment := fixDeploymentWithImage(image)
		result := converter.ToGQL(deployment)

		assert.Equal(t, expected, result)
	})

	t.Run("PR version", func(t *testing.T) {
		image := "eu.gcr.io/test/PR-1234"
		expected := gqlschema.VersionInfo{
			KymaVersion: "pull request PR-1234",
		}

		converter := &versionInfoConverter{}
		deployment := fixDeploymentWithImage(image)
		result := converter.ToGQL(deployment)

		assert.Equal(t, expected, result)
	})

	t.Run("Master version", func(t *testing.T) {
		image := "eu.gcr.io/test/12345678"
		expected := gqlschema.VersionInfo{
			KymaVersion: "master 12345678",
		}

		converter := &versionInfoConverter{}
		deployment := fixDeploymentWithImage(image)
		result := converter.ToGQL(deployment)

		assert.Equal(t, expected, result)
	})
}

func fixDeploymentWithImage(image string) *apps.Deployment {
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
							Image: image,
						},
					},
				},
			},
		},
	}
}

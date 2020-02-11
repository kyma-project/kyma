package k8s

import (
	"github.com/stretchr/testify/assert"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestDeploymentConverter_ToKymaVersion(t *testing.T) {
	t.Run("Non eu.gcr.io version", func(t *testing.T) {
		image := "test-repo/test-image"

		converter := &kymaVersionConverter{}
		deployment := fixDeploymentWithImage(image)
		result := converter.ToKymaVersion(deployment)

		assert.Equal(t, image, result)
	})

	t.Run("Scemantic version", func(t *testing.T) {
		image := "eu.gcr.io/test/1.2.3"
		converter := &kymaVersionConverter{}
		deployment := fixDeploymentWithImage(image)
		result := converter.ToKymaVersion(deployment)

		assert.Equal(t, "1.2.3", result)
	})

	t.Run("PR version", func(t *testing.T) {
		image := "eu.gcr.io/test/PR-1234"
		converter := &kymaVersionConverter{}
		deployment := fixDeploymentWithImage(image)
		result := converter.ToKymaVersion(deployment)

		assert.Equal(t, "pull request PR-1234", result)
	})

	t.Run("Master version", func(t *testing.T) {
		image := "eu.gcr.io/test/12345678"
		converter := &kymaVersionConverter{}
		deployment := fixDeploymentWithImage(image)
		result := converter.ToKymaVersion(deployment)

		assert.Equal(t, "master 12345678", result)
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
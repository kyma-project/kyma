package k8s

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeploymentConverter_ToKymaVersion(t *testing.T) {
	t.Run("Non eu.gcr.io version", func(t *testing.T) {
		image := "test-repo/test-image"

		converter := &kymaVersionConverter{}
		result := converter.ToKymaVersion(image)

		assert.Equal(t, image, result)
	})

	t.Run("Scemantic version", func(t *testing.T) {
		image := "eu.gcr.io/test/1.2.3"
		converter := &kymaVersionConverter{}
		result := converter.ToKymaVersion(image)

		assert.Equal(t, "1.2.3", result)
	})

	t.Run("PR version", func(t *testing.T) {
		image := "eu.gcr.io/test/PR-1234"
		converter := &kymaVersionConverter{}
		result := converter.ToKymaVersion(image)

		assert.Equal(t, "pull request PR-1234", result)
	})

	t.Run("Master version", func(t *testing.T) {
		image := "eu.gcr.io/test/12345678"
		converter := &kymaVersionConverter{}
		result := converter.ToKymaVersion(image)

		assert.Equal(t, "master 12345678", result)
	})
}
package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/stretchr/testify/assert"
)

func TestVersionInfoConverter_ToGQL(t *testing.T) {
	t.Run("Non eu.gcr.io version", func(t *testing.T) {
		image := "test-repo/test-image"
		expected := gqlschema.VersionInfo{
			KymaVersion: image,
		}

		converter := &versionInfoConverter{}
		deployment := fixDeployment(image)
		result := converter.ToGQL(deployment)

		assert.Equal(t, expected, result)
	})

	t.Run("Scemantic version", func(t *testing.T) {
		image := "eu.gcr.io/test/1.2.3"
		expected := gqlschema.VersionInfo{
			KymaVersion: "1.2.3",
		}

		converter := &versionInfoConverter{}
		deployment := fixDeployment(image)
		result := converter.ToGQL(deployment)

		assert.Equal(t, expected, result)
	})

	t.Run("PR version", func(t *testing.T) {
		image := "eu.gcr.io/test/PR-1234"
		expected := gqlschema.VersionInfo{
			KymaVersion: "pull request PR-1234",
		}

		converter := &versionInfoConverter{}
		deployment := fixDeployment(image)
		result := converter.ToGQL(deployment)

		assert.Equal(t, expected, result)
	})

	t.Run("Master version", func(t *testing.T) {
		image := "eu.gcr.io/test/12345678"
		expected := gqlschema.VersionInfo{
			KymaVersion: "12345678",
		}

		converter := &versionInfoConverter{}
		deployment := fixDeployment(image)
		result := converter.ToGQL(deployment)

		assert.Equal(t, expected, result)
	})
}

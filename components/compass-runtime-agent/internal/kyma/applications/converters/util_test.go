package converters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceDefinitionService_NameNormalization(t *testing.T) {

	t.Run("should replace upper case with lower case", func(t *testing.T) {
		// when
		normalised := createServiceName("DisplayName", "id")

		// then
		assert.Equal(t, "displayname-87ea5", normalised)
	})

	t.Run("should replace non alpha numeric characters with --", func(t *testing.T) {
		// when
		normalised := createServiceName("display_!@#$%^&*()name", "id")

		// then
		assert.Equal(t, "display-name-87ea5", normalised)
	})

	t.Run("should remove leading dashes", func(t *testing.T) {
		// when
		normalised := createServiceName("-----displayname", "id")

		// then
		assert.Equal(t, "displayname-87ea5", normalised)
	})

	t.Run("should trim if name too long", func(t *testing.T) {
		// when
		normalised := createServiceName("VeryVeryVeryVeryVeryVeryVEryVeryVeryVeryVeryVeryVeryVeryLongDescription", "id")

		// then
		assert.Equal(t, "veryveryveryveryveryveryveryveryveryveryveryveryveryveryl-87ea5", normalised)
	})
}

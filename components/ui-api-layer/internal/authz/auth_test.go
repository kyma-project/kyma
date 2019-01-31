package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareAttributes(t *testing.T) {

	t.Run("When no arg is required", func(t *testing.T) {

		gqlAttributes := noArgsAttributes()
		authAttributes := PrepareAttributes(noArgsContext(), &userInfo, gqlAttributes)

		verifyCommonAttributes(t, authAttributes)

		t.Run("Then namespace is empty", func(t *testing.T) {
			assert.Empty(t, authAttributes.GetNamespace())
		})
		t.Run("Then name is empty", func(t *testing.T) {
			assert.Empty(t, authAttributes.GetName())
		})

	})

	t.Run("When args are required", func(t *testing.T) {
		gqlAttributes := withArgsAttributes()
		authAttributes := PrepareAttributes(withArgsContext(), &userInfo, gqlAttributes)

		verifyCommonAttributes(t, authAttributes)

		t.Run("Then namespace is set", func(t *testing.T) {
			assert.Equal(t, namespace, authAttributes.GetNamespace())
		})
		t.Run("Then name is set", func(t *testing.T) {
			assert.Equal(t, name, authAttributes.GetName())
		})
	})
}

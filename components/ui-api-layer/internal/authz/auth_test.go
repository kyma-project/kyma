package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type exampleType struct{}

func TestPrepareAttributes(t *testing.T) {

	t.Run("When no arg is required", func(t *testing.T) {

		gqlAttributes := noArgsAttributes()
		authAttributes, err := PrepareAttributes(noArgsContext(), &userInfo, gqlAttributes, exampleType{})

		verifyCommonAttributes(t, authAttributes)

		t.Run("Then no error occured", func(t *testing.T) {
			assert.Nil(t, err)
		})
		t.Run("Then namespace is empty", func(t *testing.T) {
			assert.Empty(t, authAttributes.GetNamespace())
		})
		t.Run("Then name is empty", func(t *testing.T) {
			assert.Empty(t, authAttributes.GetName())
		})

	})

	t.Run("When args are required", func(t *testing.T) {
		gqlAttributes := withArgsAttributes()
		authAttributes, err := PrepareAttributes(withArgsContext(), &userInfo, gqlAttributes, exampleType{})

		verifyCommonAttributes(t, authAttributes)

		t.Run("Then no error occured", func(t *testing.T) {
			assert.Nil(t, err)
		})
		t.Run("Then namespace is set", func(t *testing.T) {
			assert.Equal(t, namespace, authAttributes.GetNamespace())
		})
		t.Run("Then name is set", func(t *testing.T) {
			assert.Equal(t, name, authAttributes.GetName())
		})
	})
}

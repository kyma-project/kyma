package authz

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

type exampleType struct {
	MyNamespace string
}

func TestPrepareAttributes(t *testing.T) {

	t.Run("When no arg is required", func(t *testing.T) {

		gqlAttributes := noArgsAttributes(withoutChildResolverSet)
		authAttributes, err := PrepareAttributes(noArgsContext(), &userInfo, gqlAttributes, exampleType{})

		t.Run("Then no error occured", func(t *testing.T) {
			require.Nil(t, err)
		})

		verifyCommonAttributes(t, authAttributes)

		t.Run("Then namespace is empty", func(t *testing.T) {
			assert.Empty(t, authAttributes.GetNamespace())
		})
		t.Run("Then name is empty", func(t *testing.T) {
			assert.Empty(t, authAttributes.GetName())
		})

	})

	t.Run("When arguments like name or namespace are required", func(t *testing.T) {
		t.Run("resolver is not a child resolver", func(t *testing.T) {

			gqlAttributes := withArgsAttributes(withoutChildResolverSet)

			t.Run("and arguments are provided", func(t *testing.T) {
				authAttributes, err := PrepareAttributes(withArgsContext(), &userInfo, gqlAttributes, exampleType{})

				t.Run("Then no error occured", func(t *testing.T) {
					require.Nil(t, err)
				})

				verifyCommonAttributes(t, authAttributes)

				t.Run("Then namespace is set", func(t *testing.T) {
					assert.Equal(t, namespace, authAttributes.GetNamespace())
				})
				t.Run("Then name is set", func(t *testing.T) {
					assert.Equal(t, name, authAttributes.GetName())
				})
			})

			t.Run("and arguments are not provided", func(t *testing.T) {
				authAttributes, err := PrepareAttributes(noArgsContext(), &userInfo, gqlAttributes, exampleType{})

				t.Run("Then error should be returned", func(t *testing.T) {
					require.Error(t, err)
				})
				t.Run("Then authAttributes is empty", func(t *testing.T) {
					assert.Nil(t, authAttributes)
				})
			})
		})

		t.Run("resolver is a child resolver", func(t *testing.T) {

			gqlAttributes := withNamespaceArgAttributes(withChildResolverSet)

			t.Run("and arguments in parent exists", func(t *testing.T) {
				authAttributes, err := PrepareAttributes(noArgsContext(), &userInfo, gqlAttributes, exampleType{namespace})

				t.Run("Then no error should be returned", func(t *testing.T) {
					assert.Nil(t, err)
				})

				verifyCommonAttributes(t, authAttributes)

				t.Run("Then namespace should be set", func(t *testing.T) {
					assert.Equal(t, namespace, authAttributes.GetNamespace())
				})
			})

			t.Run("and no arguments in parent exists", func(t *testing.T) {
				authAttributes, err := PrepareAttributes(noArgsContext(), &userInfo, gqlAttributes, exampleType{})

				t.Run("Then error should be returned", func(t *testing.T) {
					assert.Error(t, err)
				})
				t.Run("Then authAttributes is empty", func(t *testing.T) {
					assert.Nil(t, authAttributes)
				})
			})
		})
	})
}

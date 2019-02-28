package revocationlist

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRevocationListRepository_Insert(t *testing.T) {

	t.Run("should insert value to the list", func(t *testing.T) {
		// given
		repository := NewRepository()
		someHash := "someHash"

		// when
		err := repository.Insert(someHash)
		require.NoError(t, err)

		isPresent, err := repository.Contains(someHash)
		require.NoError(t, err)

		// then
		assert.Equal(t, isPresent, true)
	})

	t.Run("should be idempotent", func(t *testing.T) {
		// given
		repository := NewRepository()
		someHash := "someHash"

		// when
		err := repository.Insert(someHash)
		require.NoError(t, err)

		err = repository.Insert(someHash)
		require.NoError(t, err)

		isPresent, err := repository.Contains(someHash)
		require.NoError(t, err)

		// then
		assert.Equal(t, isPresent, true)
	})
}

func TestRevocationListRepository_Contains(t *testing.T) {

	t.Run("should return false if value is not present", func(t *testing.T) {
		// given
		repository := NewRepository()
		someHash := "someHash"

		// when
		isPresent, err := repository.Contains(someHash)
		require.NoError(t, err)

		// then
		assert.Equal(t, isPresent, false)
	})

	t.Run("should return true if value is present", func(t *testing.T) {
		// given
		repository := NewRepository()
		someHash := "someHash"

		repository.Insert(someHash)

		// when
		isPresent, err := repository.Contains(someHash)
		require.NoError(t, err)

		// then
		assert.Equal(t, isPresent, true)
	})
}

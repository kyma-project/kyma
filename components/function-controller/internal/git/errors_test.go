package git

import (
	"testing"

	git2go "github.com/libgit2/git2go/v34"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestIsNotRecoverableError(t *testing.T) {
	t.Run("Error is unrecoverable", func(t *testing.T) {
		//GIVEN
		err := git2go.MakeGitError2(int(git2go.ErrorCodeNotFound))
		err = errors.Wrap(err, "first")
		err = errors.Wrap(err, "second")
		err = errors.Wrap(err, "third")
		//WHEN

		res := IsNotRecoverableError(err)

		//THEN
		require.True(t, res)
	})

	t.Run("Generic recoverable error", func(t *testing.T) {
		//GIVEN
		err := errors.New("first")
		err = errors.Wrap(err, "second")
		err = errors.Wrap(err, "third")

		//WHEN
		res := IsNotRecoverableError(err)

		//THEN
		require.False(t, res)
	})
}

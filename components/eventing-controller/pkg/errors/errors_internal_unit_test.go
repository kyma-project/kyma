package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var errInvalidStorageType = NewArgumentError("invalid stream storage type: %q")

func Test_ArgumentError_WithArg(t *testing.T) {
	t.Parallel()
	// given
	errorWithoutArg := errInvalidStorageType
	someArgument := "some argument"

	// when
	errorWithArg := errInvalidStorageType.WithArg(someArgument)

	// then
	// ensure errors are still equal, otherwise Is() would return false.
	assert.Equal(t, errorWithoutArg.errorFormatType, errorWithArg.errorFormatType)
	// ensure argument is set
	assert.Equal(t, errorWithArg.argument, someArgument)
}

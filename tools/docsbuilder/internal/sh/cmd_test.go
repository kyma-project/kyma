package sh_test

import (
	"testing"

	"github.com/kyma-project/kyma/tools/docsbuilder/internal/sh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	out, err := sh.Run("echo 'Test'")
	require.NoError(t, err)
	assert.Equal(t, "Test\n", out)
}

func TestInDirRun(t *testing.T) {
	out, err := sh.RunInDir("pwd", "/")
	require.NoError(t, err)
	assert.Equal(t, "/\n", out)
}

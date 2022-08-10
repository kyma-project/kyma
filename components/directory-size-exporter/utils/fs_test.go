package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteMockFileToDirectory(t *testing.T) {
	dirPath := t.TempDir()
	file, err := WriteMockFileToDirectory(dirPath, "test.txt", 300)
	require.NoError(t, err)
	require.NotNil(t, file)
	require.Contains(t, file.Name(), "test.txt")
	fileInformation, err := os.Stat(dirPath + "/test.txt")
	require.NoError(t, err)
	require.Equal(t, fileInformation.Size(), int64(300))
}

func TestPrepareMockDirectory(t *testing.T) {
	dirPath := t.TempDir()
	err := PrepareMockDirectory(dirPath, "mockdir", 500)
	require.NoError(t, err)
	dirInformation, err := os.Stat(dirPath + "/mockdir")
	require.NoError(t, err)
	require.True(t, dirInformation.IsDir())

	file, err := os.Stat(dirPath + "/mockdir/test.txt")
	require.NoError(t, err)
	require.Equal(t, int64(500), file.Size())
}

func TestPrepareMockDirectories(t *testing.T) {
	dirPath := t.TempDir()
	mockDirPath, err := PrepareMockDirectories(dirPath)
	require.NoError(t, err)
	require.Equal(t, dirPath+"/test-data", mockDirPath)
	_, err = os.Stat(mockDirPath)
	require.NoError(t, err)

	emitter1, err := os.Stat(mockDirPath + "/emitter1")
	require.NoError(t, err)
	require.True(t, emitter1.IsDir())

	emitter2, err := os.Stat(mockDirPath + "/emitter2")
	require.NoError(t, err)
	require.True(t, emitter2.IsDir())

	emitter3, err := os.Stat(mockDirPath + "/emitter3")
	require.NoError(t, err)
	require.True(t, emitter3.IsDir())

	file, err := os.Stat(mockDirPath + "/emitter1/test.txt")
	require.NoError(t, err)
	require.Equal(t, int64(0), file.Size())

	file, err = os.Stat(mockDirPath + "/emitter2/test.txt")
	require.NoError(t, err)
	require.Equal(t, int64(100), file.Size())

	file, err = os.Stat(mockDirPath + "/emitter3/test.txt")
	require.NoError(t, err)
	require.Equal(t, int64(200), file.Size())
}

package exporter

import (
	"testing"

	"directory-size-exporter/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListDir(t *testing.T) {
	dirPath, errDirs := utils.PrepareMockDirectories(t.TempDir())
	assert.NoError(t, errDirs)

	expectedDirectories := []directory{
		{name: "emitter1", size: int64(0)},
		{name: "emitter2", size: int64(100)},
		{name: "emitter3", size: int64(200)},
	}

	directories, err := listDirs(dirPath)
	assert.NoError(t, err)

	isTrue := (len(directories) == len(expectedDirectories))
	for i, dir := range directories {
		if dir != expectedDirectories[i] {
			isTrue = false
			break
		}
	}

	require.True(t, isTrue)
}

func TestDirSize(t *testing.T) {
	dirPath, errDirs := utils.PrepareMockDirectories(t.TempDir())
	assert.NoError(t, errDirs)

	size, err := dirSize(dirPath)
	assert.NoError(t, err)

	require.Equal(t, int64(300), size)
}

func TestNewExporter(t *testing.T) {
	exporter := NewExporter("data/log", "metric_name")
	require.NotNil(t, exporter)
}

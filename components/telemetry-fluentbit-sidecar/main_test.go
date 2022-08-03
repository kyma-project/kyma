package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func prepareMockDirectories(testDir string) string {
	dirPath := testDir + "/test-data"
	err := os.Mkdir(dirPath, 0700)
	if err != nil {
		panic(err)
	}

	files := []string{"file1", "file2", "file3"}
	for i, file := range files {
		err = os.Mkdir(dirPath+"/"+file, 0700)
		if err != nil {
			panic(err)
		}
		var newFile *os.File = nil
		newFile, err = os.Create(dirPath + "/" + file + "/test.txt")
		if err != nil {
			panic(err)
		}
		err = os.Truncate(dirPath+"/"+file+"/test.txt", int64(i*100))
		if err != nil {
			panic(err)
			newFile.Close()
		}
		newFile.Close()
		if err != nil {
			panic(err)
		}
	}

	return dirPath
}

func TestListDir(t *testing.T) {
	dirPath := prepareMockDirectories(t.TempDir())

	expectedDirectories := []directory{
		{name: "file1", size: int64(0)},
		{name: "file2", size: int64(100)},
		{name: "file3", size: int64(200)},
	}

	directories, errListDirs := listDirs(dirPath)
	if errListDirs != nil {
		panic(errListDirs)
	}

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
	dirPath := prepareMockDirectories(t.TempDir())
	size, err := dirSize(dirPath)
	if err != nil {
		panic(err)
	}
	require.Equal(t, int64(300), size)
}

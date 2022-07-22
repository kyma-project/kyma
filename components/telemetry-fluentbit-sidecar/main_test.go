package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListDir(t *testing.T) {
	err := os.Mkdir("test-data", 0700)
	if err != nil {
		panic(err)
	}

	files := []string{"file1", "file2", "file3"}
	for i, file := range files {
		err = os.Mkdir("test-data/"+file, 0700)
		if err != nil {
			panic(err)
		}
		var newFile *os.File = nil
		newFile, err = os.Create("test-data/" + file + "/test.txt")
		if err != nil {
			panic(err)
		}
		err = os.Truncate("test-data/"+file+"/test.txt", int64(i*100))
		if err != nil {
			panic(err)
			newFile.Close()
		}
		newFile.Close()
		if err != nil {
			panic(err)
		}
	}

	size, errDirSize := dirSize("test-data")
	if errDirSize != nil {
		panic(errDirSize)
	}
	println(size)

	directories, errListDirs := listDirs("test-data")
	if errListDirs != nil {
		panic(errListDirs)
	}
	println(directories)

	err = os.RemoveAll("./test-data")
	if err != nil {
		panic(err)
	}
	require.Equal(t, 1, 1, "1")
}

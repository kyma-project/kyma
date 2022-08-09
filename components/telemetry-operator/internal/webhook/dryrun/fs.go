package dryrun

import (
	"os"
)

func WriteFile(path string, data string) error {
	return os.WriteFile(path, []byte(data), 0666)
}

func MakeDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func RemoveAll(path string) error {
	return os.RemoveAll(path)
}

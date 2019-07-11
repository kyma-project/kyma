package kymasources

import "os"

// FilesystemWrapper defines interface abstraction for interacting with file system
type FilesystemWrapper interface {
	Exists(path string) bool
	CreateDir(path string) error
}

// FilesystemWrapperMock is a mocked implementation of `FilesystemWrapper`` interface
type FilesystemWrapperMock struct{}

// Exists is a mocked method which always panic
func (FilesystemWrapperMock) Exists(path string) bool { panic("call") }

// CreateDir is a mocked method which always panic
func (FilesystemWrapperMock) CreateDir(path string) error { panic("call") }

type filesystemWrapper struct{}

// Exists returns true when specified path exists
func (fsw *filesystemWrapper) Exists(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err) == false
}

// CreateDir creates directory structure with specified path
func (fsw *filesystemWrapper) CreateDir(path string) error {
	return os.MkdirAll(path, os.ModePerm|os.ModeDir)
}

// NewFilesystemWrapper return instance of `FilesystemWrapper`
func NewFilesystemWrapper() FilesystemWrapper {
	return &filesystemWrapper{}
}

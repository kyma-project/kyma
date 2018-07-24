package steps

import "os"

// KymaPackageInterface .
type KymaPackageInterface interface {
	CreateDir(kymaPath string) error
	NeedDownload(kymaPath string) bool
	RemoveDir(kymaPath string) error
}

// KymaPackageClient .
type KymaPackageClient struct {
}

// NeedDownload .
func (kymaPackageClient *KymaPackageClient) NeedDownload(kymaPath string) bool {
	_, err := os.Stat(kymaPath)
	return os.IsNotExist(err)
}

// CreateDir .
func (kymaPackageClient *KymaPackageClient) CreateDir(kymaPath string) error {
	return os.MkdirAll(kymaPath, os.ModePerm|os.ModeDir)
}

// RemoveDir .
func (kymaPackageClient *KymaPackageClient) RemoveDir(kymaPath string) error {
	return os.RemoveAll(kymaPath)
}

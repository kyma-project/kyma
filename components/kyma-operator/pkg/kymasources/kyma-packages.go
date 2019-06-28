package kymasources

import (
	"errors"
	"path"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/toolkit"
)

const (
	// InjectedDirName const defines name for directory with injected kyma sources
	InjectedDirName = "injected"
)

// KymaPackages is the interface for interacting with kyma packages
type KymaPackages interface {
	HasInjectedSources() bool
	GetInjectedPackage() (KymaPackage, error)
	GetPackage(version string) (KymaPackage, error)
	FetchPackage(url, version string) error
}

// KymaPackagesMock is a mocked implementation of KymaPackages interface
type KymaPackagesMock struct{}

// HasInjectedSources is a mocked method which always panic
func (KymaPackagesMock) HasInjectedSources() bool { panic("call") }

// GetInjectedPackage is a mocked method which always panic
func (KymaPackagesMock) GetInjectedPackage() (KymaPackage, error) { panic("call") }

// GetPackage is a mocked method which always panic
func (KymaPackagesMock) GetPackage(version string) (KymaPackage, error) { panic("call") }

// FetchPackage is a mocked method which always panic
func (KymaPackagesMock) FetchPackage(url, version string) error { panic("call") }

type kymaPackages struct {
	fsWrapper   FilesystemWrapper
	cmdExecutor toolkit.CommandExecutor

	rootDir string
}

// HasInjectedSources returns true when there are kyma sources that were injected into kyma-operator
func (kps *kymaPackages) HasInjectedSources() bool {
	injectedPackagePath := kps.getInjectedPackageDirPath()
	return kps.fsWrapper.Exists(injectedPackagePath)
}

// GetInjectedPackage returns `KymaPackage` injected into kyma-operator or error when it does not exist
func (kps *kymaPackages) GetInjectedPackage() (KymaPackage, error) {
	injectedPackagePath := kps.getInjectedPackageDirPath()

	if kps.fsWrapper.Exists(injectedPackagePath) == false {
		return nil, errors.New("Unable to locate injected kyma package")
	}

	return NewKymaPackage(injectedPackagePath, "v0.0.0-injected"), nil
}

// GetPackage returns `KymaPackage` or error when it does not exist
func (kps *kymaPackages) GetPackage(version string) (KymaPackage, error) {
	packageDirPath := kps.getPackageDirPath(version)

	if kps.fsWrapper.Exists(packageDirPath) == false {
		return nil, errors.New("Unable to locate kyma package with version " + version)
	}

	return NewKymaPackage(packageDirPath, version), nil
}

// FetchPackage fetches kyma package from specified url
func (kps *kymaPackages) FetchPackage(url, version string) error {
	outputFilePath := path.Join(kps.rootDir, version+".tar.gz")
	packageDirPath := kps.getPackageDirPath(version)

	err := kps.prepareRootDir()
	if err != nil {
		return err
	}

	err = kps.downloadPackage(url, outputFilePath)
	if err != nil {
		return err
	}

	err = kps.fsWrapper.CreateDir(packageDirPath)
	if err != nil {
		return err
	}

	err = kps.extractPackage(outputFilePath, packageDirPath)
	if err != nil {
		return err
	}

	return nil
}

func (kps *kymaPackages) prepareRootDir() error {
	if kps.fsWrapper.Exists(kps.rootDir) == false {
		return kps.fsWrapper.CreateDir(kps.rootDir)
	}

	return nil
}

func (kps *kymaPackages) getInjectedPackageDirPath() string {
	injectedPackagePath := path.Join(kps.rootDir, InjectedDirName)
	return injectedPackagePath
}

func (kps *kymaPackages) getPackageDirPath(version string) string {
	packageDirPath := path.Join(kps.rootDir, version)
	return packageDirPath
}

func (kps *kymaPackages) extractPackage(packageFilePath, packageDirPath string) error {
	return kps.cmdExecutor.RunCommand("tar", "xz", "-C", packageDirPath, "--strip-components=1", "-f", packageFilePath)
}

func (kps *kymaPackages) downloadPackage(url, outputFilePath string) error {
	return kps.cmdExecutor.RunCommand("curl", "-Lks", url, "-o", outputFilePath)
}

// NewKymaPackages returns instance of `KymaPackages` type
func NewKymaPackages(fsWrapper FilesystemWrapper, cmdExecutor toolkit.CommandExecutor, rootDir string) KymaPackages {
	return &kymaPackages{
		fsWrapper:   fsWrapper,
		cmdExecutor: cmdExecutor,

		rootDir: rootDir,
	}
}

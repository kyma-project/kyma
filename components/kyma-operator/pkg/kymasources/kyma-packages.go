package kymasources

import (
	"errors"
	"path"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/toolkit"
)

const (
	// bundledDirName const defines name for directory with bundled kyma sources
	bundledDirName    = "injected"
	legacyDownloadDir = "legacy-deprecated" //This entire mechanism will soon get removed. Ensure it will not interfere with source-per-component logic.
)

// KymaPackages is the interface for interacting with kyma packages
type KymaPackages interface {
	HasBundledSources() bool
	GetBundledPackage() (KymaPackage, error)
	//Deprecated. User can specify external source for every component.
	FetchPackage(url, version string) error
	//Deprecated. User can specify external source for every component.
	GetPackage(version string) (KymaPackage, error)
}

// KymaPackagesMock is a mocked implementation of KymaPackages interface
type KymaPackagesMock struct{}

// HasBundledSources is a mocked method which always panic
func (KymaPackagesMock) HasBundledSources() bool { panic("call") }

// GetBundledPackage is a mocked method which always panic
func (KymaPackagesMock) GetBundledPackage() (KymaPackage, error) { panic("call") }

// GetPackage is a mocked method which always panic
func (KymaPackagesMock) GetPackage(version string) (KymaPackage, error) { panic("call") }

// FetchPackage is a mocked method which always panic
func (KymaPackagesMock) FetchPackage(url, version string) error { panic("call") }

type kymaPackages struct {
	fsWrapper   FilesystemWrapper
	cmdExecutor toolkit.CommandExecutor

	rootDir string
}

// HasBundledSources returns true when there are kyma sources that are bundled with Kyma-operator
func (kps *kymaPackages) HasBundledSources() bool {
	bundledPackagePath := kps.getBundledPackageDirPath()
	return kps.fsWrapper.Exists(bundledPackagePath)
}

// GetBundledPackage returns `KymaPackage` bundled with Kyma-operator or error when it does not exist
func (kps *kymaPackages) GetBundledPackage() (KymaPackage, error) {
	bundledPackagePath := kps.getBundledPackageDirPath()

	if kps.fsWrapper.Exists(bundledPackagePath) == false {
		return nil, errors.New("Unable to locate bundled kyma package")
	}

	return NewKymaPackage(bundledPackagePath, "v0.0.0-injected"), nil
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
	packageDirPath := kps.getPackageDirPath(version)

	if kps.fsWrapper.Exists(packageDirPath) {
		//Already fetched. Do not re-download
		return nil
	}
	outputFilePath := packageDirPath + ".tar.gz"

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

func (kps *kymaPackages) getBundledPackageDirPath() string {
	bundledPackagePath := path.Join(kps.rootDir, bundledDirName)
	return bundledPackagePath
}

func (kps *kymaPackages) getPackageDirPath(version string) string {
	packageDirPath := path.Join(kps.rootDir, legacyDownloadDir, version)
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

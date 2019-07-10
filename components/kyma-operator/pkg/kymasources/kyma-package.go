package kymasources

import "path"

const (
	// ChartsDirName const defines directory name which contains helm charts
	ChartsDirName = "resources"
)

// KymaPackage is the interface for interacting with kyma sources package
type KymaPackage interface {
	GetChartsDirPath() string
}

// KymaPackageMock is a mocked implementation of KymaPackage interface
type KymaPackageMock struct{}

// GetChartsDirPath is a mocked method which always panic
func (KymaPackageMock) GetChartsDirPath() string { panic("called") }

type kymaPackage struct {
	packageDirPath string
	version        string
}

// GetChartsDirPath func returns full path to helm charts directory
func (kp *kymaPackage) GetChartsDirPath() string {
	chartDirPath := path.Join(kp.packageDirPath, ChartsDirName)
	return chartDirPath
}

// NewKymaPackage returns instance of `KymaPackage` type
func NewKymaPackage(packageDirPath, version string) KymaPackage {
	return &kymaPackage{
		packageDirPath: packageDirPath,
		version:        version,
	}
}

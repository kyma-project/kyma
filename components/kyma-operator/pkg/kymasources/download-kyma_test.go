package kymasources

import (
	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDownloadKyma(t *testing.T) {
	Convey("EnsureKymaSources function", t, func() {

		Convey("should download kyma package in case of remote installation", func() {
			testInst := &config.InstallationData{
				URL:         "doesnotexists",
				KymaVersion: "test",
			}

			mockKymaPackages := &mockKymaPackagesForDownload{}

			defaultSrc := newDefaultSources(mockKymaPackages)

			kymaPackage, err := defaultSrc.ensureDefaultSources(testInst.URL, testInst.KymaVersion)

			So(err, ShouldBeNil)
			So(kymaPackage, ShouldNotBeNil)
		})

		Convey("should not download kyma package in case of local installation", func() {
			testInst := &config.InstallationData{}

			mockKymaPackages := &mockKymaPackagesForBundled{}

			defaultSrc := newDefaultSources(mockKymaPackages)
			kymaPackage, err := defaultSrc.ensureDefaultSources(testInst.URL, testInst.KymaVersion)

			So(err, ShouldBeNil)
			So(kymaPackage, ShouldNotBeNil)
		})

		Convey("should return error if url is not set", func() {
			testInst := &config.InstallationData{}

			mockKymaPackages := &mockKymaPackagesForDownload{}

			defaultSrc := newDefaultSources(mockKymaPackages)
			kymaPackage, err := defaultSrc.ensureDefaultSources(testInst.URL, testInst.KymaVersion)

			So(err, ShouldNotBeNil)
			So(kymaPackage, ShouldBeNil)
		})

	})
}

type mockKymaPackagesForBundled struct {
	KymaPackagesMock
}

func (mockKymaPackagesForBundled) HasBundledSources() bool { return true }
func (fb mockKymaPackagesForBundled) GetBundledPackage() (KymaPackage, error) {
	return NewKymaPackage("/kyma/injected", "v.0.0.0-injected"), nil
}

type mockKymaPackagesForDownload struct {
	KymaPackagesMock
}

func (mockKymaPackagesForDownload) HasBundledSources() bool                { return false }
func (mockKymaPackagesForDownload) FetchPackage(url, version string) error { return nil }
func (mockKymaPackagesForDownload) GetPackage(version string) (KymaPackage, error) {
	return NewKymaPackage("/kymapackage", version), nil
}

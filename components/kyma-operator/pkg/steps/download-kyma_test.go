package steps

import (
	"testing"
	"time"

	fake "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned/fake"
	installationInformers "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	statusmanager "github.com/kyma-project/kyma/components/kyma-operator/pkg/statusmanager"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDownloadKyma(t *testing.T) {

	Convey("EnsureKymaSources function", t, func() {

		Convey("should download kyma package in case of remote installation", func() {
			testInst := &config.InstallationData{
				URL:         "doesnotexists",
				KymaVersion: "test",
			}

			fakeClient := fake.NewSimpleClientset()
			informers := installationInformers.NewSharedInformerFactory(fakeClient, time.Second*0)
			mockStatusManager := statusmanager.NewKymaStatusManager(fakeClient, informers.Installer().V1alpha1().Installations().Lister())
			mockKymaPackages := &mockKymaPackagesForDownload{}

			kymaTestSteps := New(nil, nil, nil, mockStatusManager, nil, nil, mockKymaPackages)

			kymaPackage, err := kymaTestSteps.EnsureKymaSources(testInst)

			So(err, ShouldBeNil)
			So(kymaPackage, ShouldNotBeNil)
		})

		Convey("should not download kyma package in case of local installation", func() {
			testInst := &config.InstallationData{}

			fakeClient := fake.NewSimpleClientset()
			informers := installationInformers.NewSharedInformerFactory(fakeClient, time.Second*0)
			mockStatusManager := statusmanager.NewKymaStatusManager(fakeClient, informers.Installer().V1alpha1().Installations().Lister())
			mockKymaPackages := &mockKymaPackagesForInjected{}

			kymaTestSteps := New(nil, nil, nil, mockStatusManager, nil, nil, mockKymaPackages)

			kymaPackage, err := kymaTestSteps.EnsureKymaSources(testInst)

			So(err, ShouldBeNil)
			So(kymaPackage, ShouldNotBeNil)
		})

		Convey("should return error if url is not set", func() {
			testInst := &config.InstallationData{}

			fakeClient := fake.NewSimpleClientset()
			informers := installationInformers.NewSharedInformerFactory(fakeClient, time.Second*0)
			mockStatusManager := statusmanager.NewKymaStatusManager(fakeClient, informers.Installer().V1alpha1().Installations().Lister())

			mockKymaPackages := &mockKymaPackagesForDownload{}
			kymaTestSteps := New(nil, nil, nil, mockStatusManager, nil, nil, mockKymaPackages)

			kymaPackage, err := kymaTestSteps.EnsureKymaSources(testInst)

			So(err, ShouldNotBeNil)
			So(kymaPackage, ShouldBeNil)
		})

	})
}

type mockKymaPackagesForInjected struct {
	kymasources.KymaPackagesMock
}

func (mockKymaPackagesForInjected) HasInjectedSources() bool { return true }
func (mockKymaPackagesForInjected) GetInjectedPackage() (kymasources.KymaPackage, error) {
	return kymasources.NewKymaPackage("/injected", "v.0.0.0-injected"), nil
}

type mockKymaPackagesForDownload struct {
	kymasources.KymaPackagesMock
}

func (mockKymaPackagesForDownload) HasInjectedSources() bool               { return false }
func (mockKymaPackagesForDownload) FetchPackage(url, version string) error { return nil }
func (mockKymaPackagesForDownload) GetPackage(version string) (kymasources.KymaPackage, error) {
	return kymasources.NewKymaPackage("/kymapackage", version), nil
}

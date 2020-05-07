package kymainstallation

import (
	"errors"
	"fmt"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"

	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

func TestInstallStep(t *testing.T) {
	Convey("Run method of the", t, func() {
		Convey("install step should", func() {
			Convey("delete failed release if it is deletable", func() {
				//given
				expectedError := fmt.Sprintf("Helm install error: %s ", "failed to install release")

				mockHelmClient := &mockHelmClient{
					failInstallingRelease: true,
					isReleaseDeletable:    true,
				}

				testInstallStep := getInstallStep(mockHelmClient)

				//when
				err := testInstallStep.Run()

				//then
				So(err.Error(), ShouldEqual, expectedError)
				So(mockHelmClient.deleteReleaseCalled, ShouldBeTrue)
			})
			Convey("not delete failed release if it is not deletable", func() {
				//given
				expectedError := fmt.Sprintf("Helm install error: %s ", "failed to install release")

				mockHelmClient := &mockHelmClient{
					failInstallingRelease: true,
					isReleaseDeletable:    false,
				}

				testInstallStep := getInstallStep(mockHelmClient)

				//when
				err := testInstallStep.Run()

				//then
				So(err.Error(), ShouldEqual, expectedError)
				So(mockHelmClient.deleteReleaseCalled, ShouldBeFalse)
			})
			Convey("return an error when IsReleaseDeletable returns an error", func() {
				//given
				installError := fmt.Sprintf("Helm install error: %s ", "failed to install release")
				isDeletableError := fmt.Sprintf("Checking status of %s failed with an error: %s", "", "failed to get release status")
				expectedError := fmt.Sprintf("%s \n %s \n", installError, isDeletableError)

				mockHelmClient := &mockHelmClient{
					failInstallingRelease:  true,
					failIsReleaseDeletable: true,
				}
				testInstallStep := getInstallStep(mockHelmClient)

				//when
				err := testInstallStep.Run()

				//then
				So(err.Error(), ShouldEqual, expectedError)

			})
		})

	})

}

// Helm Client Mock

type mockHelmClient struct {
	kymahelm.ClientInterface
	failInstallingRelease  bool
	failDeletingRelease    bool
	failIsReleaseDeletable bool
	isReleaseDeletable     bool
	deleteReleaseCalled    bool
}

func (hc *mockHelmClient) IsReleaseDeletable(rname string) (bool, error) {
	if hc.failIsReleaseDeletable {
		return false, errors.New("failed to get release status")
	}
	return hc.isReleaseDeletable, nil
}

func (hc *mockHelmClient) InstallRelease(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error) {
	if hc.failInstallingRelease == true {
		err := errors.New("failed to install release")
		return nil, err
	}
	return &rls.InstallReleaseResponse{}, nil
}

func (hc *mockHelmClient) DeleteRelease(releaseName string) (*rls.UninstallReleaseResponse, error) {
	hc.deleteReleaseCalled = true
	if hc.failDeletingRelease {
		err := errors.New("failed to delete release")
		return nil, err
	}
	return &rls.UninstallReleaseResponse{}, nil
}

func (hc *mockHelmClient) ListReleases() (*rls.ListReleasesResponse, error) {
	return nil, nil
}

func (hc *mockHelmClient) ReleaseStatus(rname string) (string, error) {
	return "", nil
}

func (hc *mockHelmClient) InstallReleaseFromChart(chartdir, ns, releaseName, overrides string) (*rls.InstallReleaseResponse, error) {
	return nil, nil
}

func (hc *mockHelmClient) InstallReleaseWithoutWait(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error) {
	return nil, nil
}

func (hc *mockHelmClient) UpgradeRelease(chartDir, releaseName, overrides string) (*rls.UpdateReleaseResponse, error) {
	return nil, nil
}

func (hc *mockHelmClient) PrintRelease(release *release.Release) {}

// SourceGetter Mock

type mockSourceGetter struct {
	kymasources.SourceGetter
}

func (sg *mockSourceGetter) SrcDirFor(component v1alpha1.KymaComponent) (string, error) {
	return "testDir/testChart", nil
}

// OverrideData Mock

type mockOverrideData struct {
	overrides.OverrideData
}

func (mod *mockOverrideData) ForRelease(releaseName string) (string, error) {
	return "", nil
}

//instancja install stepa
func getInstallStep(hc *mockHelmClient) *installStep {
	return &installStep{
		step: step{
			helmClient: hc,
			component:  v1alpha1.KymaComponent{},
		},
		sourceGetter: &mockSourceGetter{},
		overrideData: &mockOverrideData{},
	}
}

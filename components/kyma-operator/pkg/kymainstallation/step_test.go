package kymainstallation

import (
	"errors"
	"fmt"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"

	"testing"

	v1alpha1 "github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

func TestInstallStep(t *testing.T) {
	Convey("Run method of the installStep should delete failed release", t, func() {
		//installStep
		expectedError := fmt.Sprintf("Helm install error: %s ", "failed to install release")

		mockHelmClient := &mockHelmClient{
			failInstallingRelease: true,
			isReleaseDeletable:    true,
		}

		testInstallStep := getInstallStep(mockHelmClient)

		err := testInstallStep.Run()

		So(err.Error(), ShouldEqual, expectedError)
		So(mockHelmClient.deleteReleaseCalled, ShouldBeTrue)
	})
}

// Helm Client Mock

type mockHelmClient struct {
	kymahelm.ClientInterface
	failInstallingRelease bool
	failDeletingRelease   bool
	isReleaseDeletable    bool
	deleteReleaseCalled   bool
}

func (hc *mockHelmClient) IsReleaseDeletable(rname string) (bool, error) {
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

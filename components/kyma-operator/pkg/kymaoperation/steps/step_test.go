package steps

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

const (
	testComponentName      = "test-component"
	testComponentNamespace = "test-namespace"
	testReleaseName        = "test-release"
)

func TestSteps(t *testing.T) {
	Convey("Run method of the", t, func() {
		Convey("upgrade step should", func() {
			Convey("print itself", func() {
				//given
				mockHelmClient := &mockHelmClient{}
				testUpgradeStep := fakeUpgradeStep(mockHelmClient)

				//when
				asString := testUpgradeStep.ToString()

				expected := "Component: test-component, Release: test-release, Namespace: test-namespace"

				//then
				So(asString, ShouldEqual, expected)
			})
			Convey("upgrade release without errors", func() {
				//given
				mockHelmClient := &mockHelmClient{}

				testUpgradeStep := fakeUpgradeStep(mockHelmClient)

				//when
				err := testUpgradeStep.Run()

				//then
				So(err, ShouldBeNil)
			})
			Convey("rollback failed upgrade", func() {
				//given
				upgradeError := fmt.Sprintf("Helm upgrade error: %s", "failed to upgrade release")
				expectedError := fmt.Sprintf("%s\nHelm rollback of %s was successfull", upgradeError, testReleaseName)

				mockHelmClient := &mockHelmClient{
					failUpgradingRelease: true,
				}

				testUpgradeStep := fakeUpgradeStep(mockHelmClient)

				//when
				err := testUpgradeStep.Run()

				//then
				So(err.Error(), ShouldEqual, expectedError)
				So(mockHelmClient.rollbackReleaseCalled, ShouldBeTrue)
			})
			Convey("return an error when release rollback fails", func() {
				//given
				upgradeError := fmt.Sprintf("Helm upgrade error: %s", "failed to upgrade release")
				rollbackError := fmt.Sprintf("Helm rollback of %s failed with an error: %s", testReleaseName, "failed to rollback release")
				expectedError := fmt.Sprintf("%s \n %s \n", upgradeError, rollbackError)

				mockHelmClient := &mockHelmClient{
					failUpgradingRelease: true,
					failRollback:         true,
				}

				testUpgradeStep := fakeUpgradeStep(mockHelmClient)

				//when
				err := testUpgradeStep.Run()

				//then
				So(err.Error(), ShouldEqual, expectedError)
			})

		})
		Convey("install step should", func() {
			Convey("print itself", func() {
				//given
				mockHelmClient := &mockHelmClient{}
				testInstallStep := fakeInstallStep(mockHelmClient)

				//when
				asString := testInstallStep.ToString()

				expected := "Component: test-component, Release: test-release, Namespace: test-namespace"

				//then
				So(asString, ShouldEqual, expected)
			})
			Convey("install release without errors", func() {
				//given
				mockHelmClient := &mockHelmClient{}

				testInstallStep := fakeInstallStep(mockHelmClient)

				//when
				err := testInstallStep.Run()

				//then
				So(err, ShouldBeNil)
			})
			Convey("delete failed release if it is deletable", func() {
				//given
				installError := fmt.Sprintf("Helm install error: %s", "failed to install release")
				expectedError := fmt.Sprintf("%s\nHelm delete of %s was successfull", installError, testReleaseName)

				mockHelmClient := &mockHelmClient{
					failInstallingRelease: true,
					isReleaseDeletable:    true,
				}

				testInstallStep := fakeInstallStep(mockHelmClient)

				//when
				err := testInstallStep.Run()

				//then
				So(err.Error(), ShouldEqual, expectedError)
				So(mockHelmClient.deleteReleaseCalled, ShouldBeTrue)
			})
			Convey("not delete failed release if it is not deletable", func() {
				//given
				expectedError := fmt.Sprintf("Helm install error: %s", "failed to install release")

				mockHelmClient := &mockHelmClient{
					failInstallingRelease: true,
					isReleaseDeletable:    false,
				}

				testInstallStep := fakeInstallStep(mockHelmClient)

				//when
				err := testInstallStep.Run()

				//then
				So(err.Error(), ShouldEqual, expectedError)
				So(mockHelmClient.deleteReleaseCalled, ShouldBeFalse)
			})
			Convey("return an error when getting the release status fails", func() {
				//given
				installError := fmt.Sprintf("Helm install error: %s", "failed to install release")
				isDeletableError := fmt.Sprintf("Checking status of %s failed with an error: %s", testReleaseName, "failed to get release status")
				expectedError := fmt.Sprintf("%s \n %s \n", installError, isDeletableError)

				mockHelmClient := &mockHelmClient{
					failInstallingRelease:  true,
					failIsReleaseDeletable: true,
				}
				testInstallStep := fakeInstallStep(mockHelmClient)

				//when
				err := testInstallStep.Run()

				//then
				So(err.Error(), ShouldEqual, expectedError)
			})
			Convey("return an error when release deletion fails", func() {
				//given
				installError := fmt.Sprintf("Helm install error: %s", "failed to install release")
				deletingError := fmt.Sprintf("Helm delete of %s failed with an error: %s", testReleaseName, "failed to delete release")
				expectedError := fmt.Sprintf("%s \n %s \n", installError, deletingError)

				mockHelmClient := &mockHelmClient{
					failInstallingRelease: true,
					failDeletingRelease:   true,
					isReleaseDeletable:    true,
				}

				testInstallStep := fakeInstallStep(mockHelmClient)

				//when
				err := testInstallStep.Run()

				//then
				So(err.Error(), ShouldEqual, expectedError)
			})

		})
		Convey("uninstall step should", func() {
			Convey("uninstall release without errors", func() {
				//given
				mockHelmClient := &mockHelmClient{}

				testUninstallStep := fakeUninstallStep(mockHelmClient)

				//when
				err := testUninstallStep.Run()

				//then
				So(err, ShouldBeNil)
			})
		})
		Convey("no-step should", func() {
			Convey("always succeed", func() {
				//given
				mockHelmClient := &mockHelmClient{}
				testNoStep := fakeNoStep(mockHelmClient)

				//when
				err := testNoStep.Run()

				//then
				So(err, ShouldBeNil)
			})
		})
	})

}

// Helm Client Mock
type mockHelmClient struct {
	kymahelm.ClientInterface
	failInstallingRelease   bool
	failUpgradingRelease    bool
	failDeletingRelease     bool
	failRollback            bool
	failIsReleaseDeletable  bool
	isReleaseDeletable      bool
	deleteReleaseCalled     bool
	rollbackReleaseCalled   bool
	listReleasesResponse    *rls.ListReleasesResponse
	releaseDeployedRevision func(string) (int32, error)
}

func (hc *mockHelmClient) ReleaseDeployedRevision(name string) (int32, error) {
	return hc.releaseDeployedRevision(name)
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

func (hc *mockHelmClient) RollbackRelease(releaseName string, revision int32) (*rls.RollbackReleaseResponse, error) {
	hc.rollbackReleaseCalled = true
	if hc.failRollback {
		err := errors.New("failed to rollback release")
		return nil, err
	}
	return &rls.RollbackReleaseResponse{}, nil
}

func (hc *mockHelmClient) ListReleases() (*rls.ListReleasesResponse, error) {
	return hc.listReleasesResponse, nil
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
	if hc.failUpgradingRelease == true {
		err := errors.New("failed to upgrade release")
		return nil, err
	}

	return &rls.UpdateReleaseResponse{
		Release: &release.Release{
			Name:      "testRelease",
			Namespace: "testNamespace",
			Version:   14,
			Info: &release.Info{
				Status: &release.Status{
					Code: release.Status_DEPLOYED,
				},
				Description: "",
			},
		},
	}, nil
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

func fakeComponent() v1alpha1.KymaComponent {
	return v1alpha1.KymaComponent{
		Name:        testComponentName,
		Namespace:   testComponentNamespace,
		ReleaseName: testReleaseName,
	}
}

func fakeInstallStep(hc *mockHelmClient) *installStep {
	return &installStep{
		step: step{
			helmClient: hc,
			component:  fakeComponent(),
		},
		sourceGetter: &mockSourceGetter{},
		overrideData: &mockOverrideData{},
	}
}

func fakeUpgradeStep(hc *mockHelmClient) *upgradeStep {
	return &upgradeStep{
		installStep: installStep{
			step: step{
				helmClient: hc,
				component:  fakeComponent(),
			},
			sourceGetter: &mockSourceGetter{},
			overrideData: &mockOverrideData{},
		},
		deployedRevision: 13,
	}
}

func fakeUninstallStep(hc *mockHelmClient) *uninstallStep {
	return &uninstallStep{
		step: step{
			helmClient: hc,
			component:  fakeComponent(),
		},
	}
}

func fakeNoStep(hc *mockHelmClient) *noStep {
	return &noStep{
		step: step{
			helmClient: hc,
			component:  fakeComponent(),
		},
	}
}

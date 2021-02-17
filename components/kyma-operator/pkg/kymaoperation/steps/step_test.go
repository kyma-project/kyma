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
				asString := testUpgradeStep.String()

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
				upgradeError := fmt.Sprintf("Helm upgrade of release \"%s\" failed: %s", testReleaseName, "failed to upgrade release.")
				expectedError := fmt.Sprintf("%s\nFinding last known deployed revision to rollback to.\nPerforming rollback to last known deployed revision: 2\nHelm rollback of release \"%s\" was successfull", upgradeError, testReleaseName)

				mockHelmClient := &mockHelmClient{
					failUpgradingRelease: true,
					releaseDeployedRevision: func(string) (int, error) {
						return 2, nil
					},
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
				upgradeError := fmt.Sprintf("Helm upgrade of release \"%s\" failed: %s", testReleaseName, "failed to upgrade release.")
				rollbackError := fmt.Sprintf("Finding last known deployed revision to rollback to.\nPerforming rollback to last known deployed revision: 2 \n Helm rollback of release \"%s\" failed with an error: %s", testReleaseName, "failed to rollback release ")
				expectedError := fmt.Sprintf("%s\n%s\n", upgradeError, rollbackError)

				mockHelmClient := &mockHelmClient{
					failUpgradingRelease: true,
					releaseDeployedRevision: func(string) (int, error) {
						return 2, nil
					},
					failRollback: true,
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
				asString := testInstallStep.String()

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
				installError := fmt.Sprintf("Helm installation of release \"%s\" failed: %s", testReleaseName, "failed to install release")
				deletingMsg := "Deleting release before retrying."
				expectedError := fmt.Sprintf("%s\n%s\nHelm delete of release \"%s\" was successfull", installError, deletingMsg, testReleaseName)

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
				expectedError := fmt.Sprintf("Helm installation of release \"%s\" failed: %s", testReleaseName, "failed to install release")

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
				installError := fmt.Sprintf("Helm installation of release \"%s\" failed: %s", testReleaseName, "failed to install release")
				isDeletableError := fmt.Sprintf("Checking status of release \"%s\" failed with an error: %s", testReleaseName, "failed to get release status")
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
				installError := fmt.Sprintf("Helm installation of release \"%s\" failed: failed to install release", testReleaseName)
				deletingMsg := "Deleting release before retrying."
				deletingError := fmt.Sprintf("Helm delete of release \"%s\" failed with an error: %s", testReleaseName, "failed to delete release")
				expectedError := fmt.Sprintf("%s\n%s \n %s \n", installError, deletingMsg, deletingError)

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
	failIsReleasePresent    bool
	failIsReleaseDeletable  bool
	isReleasePresent        bool
	isReleaseDeletable      bool
	deleteReleaseCalled     bool
	rollbackReleaseCalled   bool
	listReleasesResponse    []*kymahelm.Release
	releaseDeployedRevision func(string) (int, error)
}

func (hc *mockHelmClient) ReleaseDeployedRevision(nn kymahelm.NamespacedName) (int, error) {
	return hc.releaseDeployedRevision(nn.Name)
}

func (hc *mockHelmClient) IsReleasePresent(nn kymahelm.NamespacedName) (bool, error) {
	if hc.failIsReleasePresent {
		return false, errors.New("failed to get release status")
	}
	return hc.isReleasePresent, nil
}

func (hc *mockHelmClient) IsReleaseDeletable(nn kymahelm.NamespacedName) (bool, error) {
	if hc.failIsReleaseDeletable {
		return false, errors.New("failed to get release status")
	}
	return hc.isReleaseDeletable, nil
}

func (hc *mockHelmClient) InstallRelease(chartDir string, nn kymahelm.NamespacedName, values overrides.Map) (*kymahelm.Release, error) {
	if hc.failInstallingRelease == true {
		err := errors.New("failed to install release")
		return nil, err
	}
	return &kymahelm.Release{}, nil
}

func (hc *mockHelmClient) UninstallRelease(nn kymahelm.NamespacedName) error {
	hc.deleteReleaseCalled = true
	if hc.failDeletingRelease {
		return errors.New("failed to delete release")
	}
	return nil
}

func (hc *mockHelmClient) RollbackRelease(nn kymahelm.NamespacedName, revision int) error {
	hc.rollbackReleaseCalled = true
	if hc.failRollback {
		return errors.New("failed to rollback release")
	}
	return nil
}

func (hc *mockHelmClient) ListReleases() ([]*kymahelm.Release, error) {
	return hc.listReleasesResponse, nil
}

func (hc *mockHelmClient) ReleaseStatus(nn kymahelm.NamespacedName) (string, error) {
	return "", nil
}

func (hc *mockHelmClient) UpgradeRelease(chartDir string, nn kymahelm.NamespacedName, values overrides.Map) (*kymahelm.Release, error) {
	if hc.failUpgradingRelease == true {
		err := errors.New("failed to upgrade release")
		return nil, err
	}

	return &kymahelm.Release{
		ReleaseMeta: &kymahelm.ReleaseMeta{
			NamespacedName: kymahelm.NamespacedName{
				Name:      "testRelease",
				Namespace: "testNamespace",
			},
			Description: "",
		},
		ReleaseStatus: &kymahelm.ReleaseStatus{
			CurrentRevision: 14,
			Status:          kymahelm.StatusDeployed,
		}}, nil
}

func (hc *mockHelmClient) PrintRelease(release *kymahelm.Release) {}

func (hc *mockHelmClient) WaitForReleaseDelete(nn kymahelm.NamespacedName) (bool, error) {
	return true, nil
}

func (hc *mockHelmClient) WaitForReleaseRollback(nn kymahelm.NamespacedName) (bool, error) {
	return true, nil
}

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

func (mod *mockOverrideData) ForRelease(releaseName string) overrides.Map {
	return nil
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

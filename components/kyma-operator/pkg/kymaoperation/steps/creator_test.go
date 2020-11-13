package steps

import (
	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStepsFactoryCreator(t *testing.T) {
	Convey("stepFactoryCreator.getInstalledReleases()", t, func() {
		Convey("should behave correctly for nil release data", func() {
			//given
			mockHelmClient := &mockHelmClient{}
			sfc := &stepFactoryCreator{
				helmClient: mockHelmClient,
			}

			//when
			releases, err := sfc.getInstalledReleases()

			//then
			So(err, ShouldBeNil)
			So(len(releases), ShouldEqual, 0)
		})
		Convey("should return release data for deployed revisions", func() {
			//given
			mockHelmClient := &mockHelmClient{
				listReleasesResponse: fakeReleasesNoErrors(),
			}
			sfc := &stepFactoryCreator{
				helmClient: mockHelmClient,
			}

			//when
			releases, err := sfc.getInstalledReleases()

			//then
			So(err, ShouldBeNil)
			So(len(releases), ShouldEqual, 3)
			So(releases["abc"].Status, ShouldEqual, kymahelm.StatusDeployed)
			So(releases["abc"].CurrentRevision, ShouldEqual, 3)
			So(releases["abc"].LastDeployedRevision, ShouldEqual, 3)
			So(releases["ijk"].Status, ShouldEqual, kymahelm.StatusDeployed)
			So(releases["ijk"].CurrentRevision, ShouldEqual, 1)
			So(releases["ijk"].LastDeployedRevision, ShouldEqual, 1)
			So(releases["xyz"].Status, ShouldEqual, kymahelm.StatusDeployed)
			So(releases["xyz"].CurrentRevision, ShouldEqual, 2)
			So(releases["xyz"].LastDeployedRevision, ShouldEqual, 2)
		})
		Convey("should return release data for revisions with some failures", func() {
			//given
			listReleasesResponse, releaseDeployedRevision := fakeReleasesWithErrors()
			mockHelmClient := &mockHelmClient{
				listReleasesResponse:    listReleasesResponse,
				releaseDeployedRevision: releaseDeployedRevision,
			}

			sfc := &stepFactoryCreator{
				helmClient: mockHelmClient,
			}

			//when
			releases, err := sfc.getInstalledReleases()

			//then
			So(err, ShouldBeNil)
			So(len(releases), ShouldEqual, 3)
			So(releases["abc"].Status, ShouldEqual, kymahelm.StatusFailed)
			So(releases["abc"].CurrentRevision, ShouldEqual, 3)
			So(releases["abc"].LastDeployedRevision, ShouldEqual, 2)
			So(releases["ijk"].Status, ShouldEqual, kymahelm.StatusUninstalled)
			So(releases["ijk"].CurrentRevision, ShouldEqual, 1)
			So(releases["ijk"].LastDeployedRevision, ShouldEqual, 0)
			So(releases["xyz"].Status, ShouldEqual, kymahelm.StatusDeployed)
			So(releases["xyz"].CurrentRevision, ShouldEqual, 2)
			So(releases["xyz"].LastDeployedRevision, ShouldEqual, 2)
		})
	})
	Convey("stepFactoryCreator.ForInstallation()", t, func() {
		Convey("should return a correct StepLister", func() {
			//given
			mockHelmClient := &mockHelmClient{
				listReleasesResponse: fakeReleasesNoErrors(),
			}
			overrides := fakeOverrides{}
			installationData := fakeInstallationData()
			sourceGetterSupport := fakeSourceGetterLegacySupport{}
			sfc := &stepFactoryCreator{
				sourceGetterSupport: sourceGetterSupport,
				helmClient:          mockHelmClient,
			}
			//when
			stepLister, err := sfc.ForInstallation(installationData, overrides)
			So(err, ShouldBeNil)

			//then
			_, ok := stepLister.(*installStepFactory)
			So(ok, ShouldBeTrue)

			//and when
			steps, err := stepLister.StepList()

			//then
			So(err, ShouldBeNil)
			So(steps, ShouldHaveLength, 3)
			So(steps[0].String(), ShouldEqual, "Component: componentOne, Release: componentOne, Namespace: testNamespaceOne")
			So(steps[1].String(), ShouldEqual, "Component: componentTwo, Release: componentTwo, Namespace: testNamespaceTwo")
			So(steps[2].String(), ShouldEqual, "Component: componentThree, Release: componentThree, Namespace: testNamespaceTwo")
		})
	})
	Convey("stepFactoryCreator.ForUninstallation()", t, func() {
		Convey("should return an instance of uninstallStepFactory", func() {
			//given
			mockHelmClient := &mockHelmClient{
				listReleasesResponse: fakeReleasesNoErrors(),
			}
			installationData := fakeInstallationData()
			sfc := &stepFactoryCreator{
				helmClient: mockHelmClient,
			}
			//when
			stepLister, err := sfc.ForUninstallation(installationData)
			So(err, ShouldBeNil)

			//then
			_, ok := stepLister.(*uninstallStepFactory)
			So(ok, ShouldBeTrue)

			//and when
			steps, err := stepLister.StepList()

			//then
			So(err, ShouldBeNil)
			So(steps, ShouldHaveLength, 3)
			So(steps[0].String(), ShouldEqual, "Component: componentOne, Release: componentOne, Namespace: testNamespaceOne")
			So(steps[1].String(), ShouldEqual, "Component: componentTwo, Release: componentTwo, Namespace: testNamespaceTwo")
			So(steps[2].String(), ShouldEqual, "Component: componentThree, Release: componentThree, Namespace: testNamespaceTwo")
		})
	})
}

func fakeRelease(name string, version int, status kymahelm.Status) *kymahelm.Release {

	return &kymahelm.Release{
		ReleaseMeta: &kymahelm.ReleaseMeta{
			NamespacedName: kymahelm.NamespacedName{Name: name, Namespace: ""},
			Description:    "",
		},
		ReleaseStatus: &kymahelm.ReleaseStatus{
			CurrentRevision: version,
			Status:          status,
		},
	}
}

func fakeReleasesNoErrors() []*kymahelm.Release {
	r1 := fakeRelease("abc", 1, kymahelm.StatusDeployed)
	r2 := fakeRelease("abc", 2, kymahelm.StatusDeployed)
	r3 := fakeRelease("abc", 3, kymahelm.StatusDeployed)
	r4 := fakeRelease("ijk", 1, kymahelm.StatusDeployed)
	r5 := fakeRelease("xyz", 1, kymahelm.StatusDeployed)
	r6 := fakeRelease("xyz", 2, kymahelm.StatusDeployed)

	return []*kymahelm.Release{r1, r2, r3, r4, r5, r6}
}

//returns a list of fake releases and a function that returns last deployed release for given release name
func fakeReleasesWithErrors() ([]*kymahelm.Release, func(string) (int, error)) {

	releaseDeployedRevision := func(releaseName string) (int, error) {

		switch releaseName {
		case "abc":
			return 2, nil
		case "ijk":
			return 0, nil
		case "xyz":
			return 2, nil
		}

		panic("Unknown release")
	}

	r1 := fakeRelease("abc", 1, kymahelm.StatusFailed)
	r2 := fakeRelease("abc", 2, kymahelm.StatusDeployed)
	r3 := fakeRelease("abc", 3, kymahelm.StatusFailed)
	r4 := fakeRelease("ijk", 1, kymahelm.StatusUninstalled)
	r5 := fakeRelease("xyz", 1, kymahelm.StatusFailed)
	r6 := fakeRelease("xyz", 2, kymahelm.StatusDeployed)

	return []*kymahelm.Release{r1, r2, r3, r4, r5, r6}, releaseDeployedRevision
}

type fakeOverrides struct {
}

func (fo fakeOverrides) ForRelease(releaseName string) overrides.Map {
	return overrides.Map{}
}

func fakeInstallationData() *config.InstallationData {

	return &config.InstallationData{
		Context:     config.InstallationContext{},
		KymaVersion: "",
		URL:         "",
		Components: []v1alpha1.KymaComponent{
			v1alpha1.KymaComponent{Name: "componentOne", Namespace: "testNamespaceOne"},
			v1alpha1.KymaComponent{Name: "componentTwo", Namespace: "testNamespaceTwo"},
			v1alpha1.KymaComponent{Name: "componentThree", Namespace: "testNamespaceTwo"},
		},
		Action:  "install",
		Profile: "test",
	}
}

type fakeSourceGetterLegacySupport struct {
}

func (fssgls fakeSourceGetterLegacySupport) SourceGetterFor(kymaURL, kymaVersion string) SourceGetter {
	return nil
}

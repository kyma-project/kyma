package kymainstallation

import (
	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStepsFactory(t *testing.T) {
	Convey("stepFactorCreator.getInstalledReleases()", t, func() {
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

	Convey("installStepFactory.NewStep", t, func() {
		Convey("should return install step for non-existing release", func() {
			//given
			isf := installStepFactory{
				stepFactory: stepFactory{
					installedReleases: nil,
				},
			}
			component := fakeComponent()

			//when
			res, err := isf.NewStep(component)

			//then
			So(err, ShouldBeNil)
			So(res.ToString(), ShouldEqual, "Component: test-component, Release: test-release, Namespace: test-namespace")
			So(res, ShouldHaveSameTypeAs, installStep{})
		})

		Convey("should return upgrade step for existing release", func() {
			//given
			component := fakeComponent()

			installedReleases := map[string]kymahelm.ReleaseStatus{}
			installedReleases[component.GetReleaseName()] = kymahelm.ReleaseStatus{
				Status:               kymahelm.StatusDeployed,
				CurrentRevision:      2,
				LastDeployedRevision: 2,
			}

			isf := installStepFactory{
				stepFactory: stepFactory{
					installedReleases: installedReleases,
				},
			}

			//when
			res, err := isf.NewStep(component)

			//then
			So(err, ShouldBeNil)
			So(res.ToString(), ShouldEqual, "Component: test-component, Release: test-release, Namespace: test-namespace")
			So(res, ShouldHaveSameTypeAs, upgradeStep{})
		})
	})
	Convey("uninstallStepFactory.NewStep", t, func() {
		Convey("should return nostep for non-existing component", func() {
			//given

			component := fakeComponent()

			usf := uninstallStepFactory{
				stepFactory: stepFactory{
					installedReleases: nil,
				},
			}

			//when
			res, err := usf.NewStep(component)

			//then
			So(err, ShouldBeNil)
			So(res.ToString(), ShouldEqual, "Component: test-component, Release: test-release, Namespace: test-namespace")
			So(res, ShouldHaveSameTypeAs, noStep{})
		})
		Convey("should return uninstall step for deployed component", func() {
			//given

			component := fakeComponent()

			installedReleases := map[string]kymahelm.ReleaseStatus{}
			installedReleases[component.GetReleaseName()] = kymahelm.ReleaseStatus{
				Status:               kymahelm.StatusDeployed,
				CurrentRevision:      2,
				LastDeployedRevision: 2,
			}

			usf := uninstallStepFactory{
				stepFactory: stepFactory{
					installedReleases: installedReleases,
				},
			}

			//when
			res, err := usf.NewStep(component)

			//then
			So(err, ShouldBeNil)
			So(res.ToString(), ShouldEqual, "Component: test-component, Release: test-release, Namespace: test-namespace")
			So(res, ShouldHaveSameTypeAs, uninstallStep{})
		})
	})
}

func fakeRelease(name string, version int, status kymahelm.Status) *kymahelm.Release {

	return &kymahelm.Release{
		ReleaseMeta: &kymahelm.ReleaseMeta{
			Name:        name,
			Description: "",
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

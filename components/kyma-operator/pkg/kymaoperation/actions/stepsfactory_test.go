package actions

import (
	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	. "github.com/smartystreets/goconvey/convey"
	release "k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
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
			So(releases["abc"].StatusCode, ShouldEqual, release.Status_DEPLOYED)
			So(releases["abc"].CurrentRevision, ShouldEqual, 3)
			So(releases["abc"].LastDeployedRevision, ShouldEqual, 3)
			So(releases["ijk"].StatusCode, ShouldEqual, release.Status_DEPLOYED)
			So(releases["ijk"].CurrentRevision, ShouldEqual, 1)
			So(releases["ijk"].LastDeployedRevision, ShouldEqual, 1)
			So(releases["xyz"].StatusCode, ShouldEqual, release.Status_DEPLOYED)
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
			So(releases["abc"].StatusCode, ShouldEqual, release.Status_FAILED)
			So(releases["abc"].CurrentRevision, ShouldEqual, 3)
			So(releases["abc"].LastDeployedRevision, ShouldEqual, 2)
			So(releases["ijk"].StatusCode, ShouldEqual, release.Status_DELETED)
			So(releases["ijk"].CurrentRevision, ShouldEqual, 1)
			So(releases["ijk"].LastDeployedRevision, ShouldEqual, 0)
			So(releases["xyz"].StatusCode, ShouldEqual, release.Status_DEPLOYED)
			So(releases["xyz"].CurrentRevision, ShouldEqual, 2)
			So(releases["xyz"].LastDeployedRevision, ShouldEqual, 2)
		})
	})

	Convey("installStepFactory.newStep", t, func() {
		Convey("should return install step for non-existing release", func() {
			//given
			isf := installStepFactory{
				stepFactory: stepFactory{
					installedReleases: nil,
				},
			}
			component := fakeComponent()

			//when
			res, err := isf.newStep(component)

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
				StatusCode:           release.Status_DEPLOYED,
				CurrentRevision:      2,
				LastDeployedRevision: 2,
			}

			isf := installStepFactory{
				stepFactory: stepFactory{
					installedReleases: installedReleases,
				},
			}

			//when
			res, err := isf.newStep(component)

			//then
			So(err, ShouldBeNil)
			So(res.ToString(), ShouldEqual, "Component: test-component, Release: test-release, Namespace: test-namespace")
			So(res, ShouldHaveSameTypeAs, upgradeStep{})
		})
	})
	Convey("uninstallStepFactory.newStep", t, func() {
		Convey("should return nostep for non-existing component", func() {
			//given

			component := fakeComponent()

			usf := uninstallStepFactory{
				stepFactory: stepFactory{
					installedReleases: nil,
				},
			}

			//when
			res, err := usf.newStep(component)

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
				StatusCode:           release.Status_DEPLOYED,
				CurrentRevision:      2,
				LastDeployedRevision: 2,
			}

			usf := uninstallStepFactory{
				stepFactory: stepFactory{
					installedReleases: installedReleases,
				},
			}

			//when
			res, err := usf.newStep(component)

			//then
			So(err, ShouldBeNil)
			So(res.ToString(), ShouldEqual, "Component: test-component, Release: test-release, Namespace: test-namespace")
			So(res, ShouldHaveSameTypeAs, uninstallStep{})
		})
	})
}

func fakeRelease(name string, version int32, code release.Status_Code) *release.Release {
	return &release.Release{
		Name:    name,
		Version: version,
		Info: &release.Info{
			Status: &release.Status{
				Code: code,
			},
		},
	}
}

func fakeReleasesNoErrors() *rls.ListReleasesResponse {
	r1 := fakeRelease("abc", 1, release.Status_DEPLOYED)
	r2 := fakeRelease("abc", 2, release.Status_DEPLOYED)
	r3 := fakeRelease("abc", 3, release.Status_DEPLOYED)
	r4 := fakeRelease("ijk", 1, release.Status_DEPLOYED)
	r5 := fakeRelease("xyz", 1, release.Status_DEPLOYED)
	r6 := fakeRelease("xyz", 2, release.Status_DEPLOYED)

	releases := []*release.Release{r1, r2, r3, r4, r5, r6}
	return &rls.ListReleasesResponse{
		Releases: releases,
	}
}

//returns a list of fake releases and a function that returns last deployed release for given release name
func fakeReleasesWithErrors() (*rls.ListReleasesResponse, func(string) (int32, error)) {

	releaseDeployedRevision := func(releaseName string) (int32, error) {

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

	r1 := fakeRelease("abc", 1, release.Status_FAILED)
	r2 := fakeRelease("abc", 2, release.Status_DEPLOYED)
	r3 := fakeRelease("abc", 3, release.Status_FAILED)
	r4 := fakeRelease("ijk", 1, release.Status_DELETED)
	r5 := fakeRelease("xyz", 1, release.Status_FAILED)
	r6 := fakeRelease("xyz", 2, release.Status_DEPLOYED)

	releases := []*release.Release{r1, r2, r3, r4, r5, r6}
	return &rls.ListReleasesResponse{
		Releases: releases,
	}, releaseDeployedRevision
}

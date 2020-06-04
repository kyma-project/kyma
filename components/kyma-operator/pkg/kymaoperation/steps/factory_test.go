package steps

import (
	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStepsFactory(t *testing.T) {
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
			So(res.String(), ShouldEqual, "Component: test-component, Release: test-release, Namespace: test-namespace")
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
			res, err := isf.newStep(component)

			//then
			So(err, ShouldBeNil)
			So(res.String(), ShouldEqual, "Component: test-component, Release: test-release, Namespace: test-namespace")
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
			So(res.String(), ShouldEqual, "Component: test-component, Release: test-release, Namespace: test-namespace")
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
			res, err := usf.newStep(component)

			//then
			So(err, ShouldBeNil)
			So(res.String(), ShouldEqual, "Component: test-component, Release: test-release, Namespace: test-namespace")
			So(res, ShouldHaveSameTypeAs, uninstallStep{})
		})
	})
}

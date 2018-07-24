package steps

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInstallCore(t *testing.T) {

	Convey("InstallCore function", t, func() {

		Convey("In case no error occurs, should", func() {

			kymaTestSteps, installationData, mockHelmClient, _ := getTestSetup()

			Convey("call UpdateInstallationStatus once and call InstalRelease, returning no error", func() {

				err := kymaTestSteps.InstallCore(installationData)

				So(mockHelmClient.InstallReleaseCalled, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})

		Convey("In case en error occurs, should", func() {

			kymaTestSteps, installationData, mockErrorHelmClient, _ := getFailingTestSetup()

			Convey("call UpdateInstallationStatus once, call InstallRelease, call UpdateInstallationStatus, call ReleaseStatus again and return the error", func() {

				err := kymaTestSteps.InstallCore(installationData)

				So(mockErrorHelmClient.InstallReleaseCalled, ShouldBeTrue)
				So(mockErrorHelmClient.ReleaseStatusCalled, ShouldBeTrue)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestUpgradeCore(t *testing.T) {

	Convey("UpgradeCore function", t, func() {

		Convey("In case no error occurs, should", func() {

			kymaTestSteps, installationData, mockHelmClient, _ := getTestSetup()

			Convey("call UpdateInstallationStatus once and call UpdateRelease, returning no error", func() {

				err := kymaTestSteps.UpgradeCore(installationData)

				So(mockHelmClient.UpgradeReleaseCalled, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})

		Convey("In case en error occurs, should", func() {

			kymaTestSteps, installationData, mockErrorHelmClient, _ := getFailingTestSetup()

			Convey("call UpdateInstallationStatus once, call UpdateRelease, call UpdateInstallationStatus, call ReleaseStatus again and return the error", func() {

				err := kymaTestSteps.UpgradeCore(installationData)

				So(mockErrorHelmClient.UpgradeReleaseCalled, ShouldBeTrue)
				So(mockErrorHelmClient.ReleaseStatusCalled, ShouldBeTrue)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

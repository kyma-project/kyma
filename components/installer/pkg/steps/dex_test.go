package steps

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInstallDex(t *testing.T) {

	Convey("InstallDex function", t, func() {

		Convey("In case no error occurs, should", func() {

			kymaTestSteps, installationData, mockHelmClient, _ := getTestSetup()

			Convey("call UpdateInstallationStatus once and call InstallRelease, returning no error", func() {

				err := kymaTestSteps.InstallDex(installationData)

				So(mockHelmClient.InstallReleaseCalled, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})

		Convey("In case an error occurs, should", func() {

			kymaTestSteps, installationData, mockErrorHelmClient, _ := getFailingTestSetup()

			Convey("call UpdateInstallationStatus once, call InstallRelease, call UpdateInstallationStatus again and return the error", func() {

				err := kymaTestSteps.InstallDex(installationData)

				So(mockErrorHelmClient.InstallReleaseCalled, ShouldBeTrue)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestUpdateDex(t *testing.T) {

	Convey("UpdateDex function", t, func() {

		Convey("In case no error occurs, should", func() {

			kymaTestSteps, installationData, mockHelmClient, _ := getTestSetup()

			Convey("call UpdateInstallationStatus once and call UpdateRelease, returning no error", func() {

				err := kymaTestSteps.UpdateDex(installationData)

				So(mockHelmClient.UpgradeReleaseCalled, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})

		Convey("In case an error occurs, should", func() {

			kymaTestSteps, installationData, mockErrorHelmClient, _ := getFailingTestSetup()

			Convey("call UpdateInstallationStatus once, call UpdateRelease, call UpdateInstallationStatus again and return the error", func() {

				err := kymaTestSteps.UpdateDex(installationData)

				So(mockErrorHelmClient.UpgradeReleaseCalled, ShouldBeTrue)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

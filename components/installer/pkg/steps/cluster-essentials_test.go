package steps

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInstallClusterEssentials(t *testing.T) {

	Convey("InstallClusterEssentials function", t, func() {

		Convey("In case no error occurs, should", func() {

			kymaTestSteps, testInst, mockHelmClient, _ := getTestSetup()

			Convey("call UpdateInstallationStatus once and call InstallRelease returning no error", func() {

				err := kymaTestSteps.InstallClusterEssentials(testInst)

				So(mockHelmClient.InstallReleaseCalled, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})

		Convey("In case an error occurs, should", func() {

			kymaTestSteps, testInst, mockErrorHelmClient, _ := getFailingTestSetup()

			Convey("call UpdateInstallationStatus once, call InstallRelease, call UpdateInstallationStatus again and return the error", func() {

				err := kymaTestSteps.InstallClusterEssentials(testInst)

				So(mockErrorHelmClient.InstallReleaseCalled, ShouldBeTrue)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

package steps

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInstallPrometheus(t *testing.T) {

	Convey("InstallPrometheus function", t, func() {

		Convey("In case no error occurs, should", func() {

			kymaTestSteps, installationData, mockHelmClient, _ := getTestSetup()

			Convey("call UpdateInstallationStatus once and call InstallRelease returning no error", func() {

				err := kymaTestSteps.InstallPrometheus(installationData)

				So(mockHelmClient.InstallReleaseCalled, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})

		Convey("In case helm client error occurs, should", func() {

			kymaTestSteps, installationData, mockErrorHelmClient, _ := getFailingTestSetup()

			Convey("call UpdateInstallationStatus once, call InstallRelease, call UpdateInstallationStatus again and return the error", func() {

				err := kymaTestSteps.InstallPrometheus(installationData)

				So(mockErrorHelmClient.InstallReleaseCalled, ShouldBeTrue)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

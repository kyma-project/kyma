package steps

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInstallTiller(t *testing.T) {

	Convey("InstallTiller function", t, func() {

		Convey("In case no error occurs, should", func() {

			kymaTestSteps, testInst, _, mockCommandExecutor := getTestSetup()

			Convey("call UpdateInstallationStatus once and call RunCommand returning no error", func() {

				err := kymaTestSteps.InstallTiller(testInst)

				So(mockCommandExecutor.TimesMockCommandExecutorCalled, ShouldEqual, 1)
				So(err, ShouldBeNil)
			})
		})

		Convey("In case an error occurs, should", func() {

			kymaTestSteps, testInst, _, mockFailingCommandExecutor := getFailingTestSetup()

			Convey("call UpdateInstallationStatus once, call RunCommand, call UpdateInstallationStatus again and return the error", func() {

				err := kymaTestSteps.InstallTiller(testInst)

				So(mockFailingCommandExecutor.MockFailingCommandExecutorCalled, ShouldBeTrue)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

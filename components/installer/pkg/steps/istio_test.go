package steps

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInstallIstio(t *testing.T) {

	Convey("InstallIstio", t, func() {

		Convey("In case no error occurs, should", func() {

			kymaTestSteps, installationData, mockHelmClient, mockCommandExecutor := getTestSetup()

			Convey("call UpdateInstallationStatus once, call RunCommand once, call InstallReleaseWithoutWait, call RunCommand again and return no error", func() {

				err := kymaTestSteps.InstallIstio(installationData)

				So(mockHelmClient.InstallReleaseWithoutWaitCalled, ShouldBeTrue)
				// We no longer execute shell commands during istio installation
				So(mockCommandExecutor.TimesMockCommandExecutorCalled, ShouldEqual, 0)
				So(err, ShouldBeNil)
			})

		})

		Convey("in case InstallRelease error occurs", func() {

			kymaTestSteps, installationData, mockErrorHelmClient, _ := getFailingTestSetup()
			mockCommandExecutor := &MockCommandExecutor{}

			Convey("call UpdateInstallationStatus once, call RunCommand once, call InstallReleaseWithoutWait, call UpdateInstallationStatus again and return the error", func() {

				err := kymaTestSteps.InstallIstio(installationData)

				// We no longer execute shell commands during istio installation
				So(mockCommandExecutor.TimesMockCommandExecutorCalled, ShouldEqual, 0)
				So(mockErrorHelmClient.InstallReleaseWithoutWaitCalled, ShouldBeTrue)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

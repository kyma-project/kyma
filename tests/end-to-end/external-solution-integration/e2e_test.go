package goconvey_e2e_example

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestApplicationCRDCreation(t *testing.T) {

	Convey("Given application data", t, func() {

		appName := "e2e-app"
		accessLabel := "a1b2c3"
		skipInstallation := false

		Convey("When CRD created", func() {

			K8SClient, err := testkit.NewK8sResourcesClient("kyma-integration")
			So(err, ShouldBeNil)

			app, err := K8SClient.CreateDummyApplication(appName, accessLabel, skipInstallation)
			So(err, ShouldBeNil)
			So(app, ShouldNotBeNil)

			Convey("The data within CRD should match Application Data", func() {

				app, err := K8SClient.GetApplication(appName, v1.GetOptions{})
				So(err, ShouldBeNil)
				So(app.Name, ShouldEqual, appName)
				So(app.Spec.AccessLabel, ShouldEqual, accessLabel)
				So(app.Spec.SkipInstallation, ShouldEqual, skipInstallation)

				time.Sleep(5 * time.Second)

				checker := testkit.NewK8sChecker(K8SClient, appName)
				err = checker.CheckK8sResources()

				So(err, ShouldBeNil)
			})

		})

	})

}

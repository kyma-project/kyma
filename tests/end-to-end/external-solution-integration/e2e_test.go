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

func TestSecureConnection(t *testing.T) {

	Convey("Given external solution", t, func() {

		Convey("When token requested", func() {

			Convey("The application should be able to fetch certificate and request service data", nil)

		})

	})

	//Convey("Given external solution", t, func() {
	//
	//	Convey("When token requested", func() {
	//
	//		Convey("The operator should insert token into TokenRequest CRD", nil)
	//
	//	})
	//
	//	Convey("When information on CSR requested", func() {
	//
	//		Convey("The Connector Service should return Cluster Info", nil)
	//
	//	})
	//
	//	Convey("When CSR sent", func() {
	//
	//		Convey("The Connector Service should return signed certificate", nil)
	//
	//	})
	//
	//	Convey("When /v1/services requested", func() {
	//
	//		Convey("It should return proper data when using signed certificate", nil)
	//
	//	})
	//
	//})

}

func TestServiceRegistration(t *testing.T) {

	Convey("Given valid service data", t, func() {

		Convey("When registration request is sent", func() {

			Convey("It should register the service", nil)

		})

	})

}

func TestEventsRegistration(t *testing.T) {

	Convey("Given valid events data", t, func() {

		Convey("When registration request is sent", func() {

			Convey("It should register the events", nil)

		})

	})

}

func TestBindings(t *testing.T) {

	Convey("Given application and environment", t, func() {

		Convey("When binding is created", func() {

			Convey("It should generate proper service classes", nil)

		})

	})

	Convey("Given service and environment", t, func() {

		Convey("When binding is created", func() {

			Convey("It should create proper service instances", nil)

		})

	})

	Convey("Given events and environment", t, func() {

		Convey("When binding is created", func() {

			Convey("It should create proper service instances", nil)

		})

	})

}

func TestLambda(t *testing.T) {

	Convey("Given a Lambda is present", t, func() {

		Convey("When event-trigger created", func() {

			Convey("It should be accessible within the system", nil)

		})

		Convey("When service binding is created", func() {

			Convey("It should be containing information on lambda's triggers", nil)

		})

	})

}

func TestEventSent(t *testing.T) {
	Convey("When event is sent", func() {

		Convey("Lambda should react to the event", nil)

			Convey("Request should be proxied and response delivered", nil)
	})
}

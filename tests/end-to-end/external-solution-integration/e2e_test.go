package goconvey_e2e_example

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

const (
	APPNAME     = "e2e-app"
	ACCESSLABEL = "a1b2c3"
	NAMESPACE   = "kyma-integration"
)

func TestApplicationCRDCreation(t *testing.T) {

	Convey("Given application data", t, func() {

		Convey("When CRD created", func() {

			K8SClient, err := testkit.NewK8sResourcesClient(NAMESPACE)
			So(err, ShouldBeNil)

			app, err := K8SClient.CreateDummyApplication(APPNAME, ACCESSLABEL, false)
			So(err, ShouldBeNil)
			So(app, ShouldNotBeNil)

			Convey("The data within CRD should match Application Data", func() {

				app, err := K8SClient.GetApplication(APPNAME, v1.GetOptions{})
				So(err, ShouldBeNil)
				So(app.Name, ShouldEqual, APPNAME)
				So(app.Spec.AccessLabel, ShouldEqual, ACCESSLABEL)
				So(app.Spec.SkipInstallation, ShouldEqual, false)

				time.Sleep(5 * time.Second)

				checker := testkit.NewK8sChecker(K8SClient, APPNAME)
				err = checker.CheckK8sResources()

				So(err, ShouldBeNil)
			})
		})
	})
}

func TestSecureConnection(t *testing.T) {

	Convey("Given external solution", t, func() {

		Convey("When token requested", func() {

			tokenRequestClient, err := testkit.NewTokenRequestClient(NAMESPACE)
			So(err, ShouldBeNil)

			tokenRequest, err := tokenRequestClient.CreateTokenRequest(APPNAME)
			So(err, ShouldBeNil)
			So(tokenRequest, ShouldNotBeNil)

			Convey("The operator should insert token into TokenRequest CRD", nil)

			//TODO: Polling of tokenRequest
			time.Sleep(5 * time.Second)

			tokenRequest, err = tokenRequestClient.GetTokenRequest(APPNAME, v1.GetOptions{})
			So(err, ShouldBeNil)
			So(tokenRequest.Status.Token, ShouldNotBeNil)

			Convey("When one-click-integration initiated", func() {

				mockAppClient := testkit.NewMockApplicationClient()

				res, err := mockAppClient.ConnectToKyma(tokenRequest.Status.URL, true, false)
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)

				Convey("The app should be connected to Kyma", func() {

					res, err = mockAppClient.GetConnectionInfo()
					So(err, ShouldBeNil)
					So(res.AppName, ShouldEqual, APPNAME)
					So(res.ClusterDomain, ShouldNotBeNil)
					So(res.EventsURL, ShouldNotBeNil)
					So(res.MetadataURL, ShouldNotBeNil)
				})
			})
		})
	})
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

func TestLambdaCreation(t *testing.T) {

	Convey("Given Lambda's code", t, func() {

		Convey("When event-trigger created", func() {

			Convey("It should be accessible within the system", nil)

		})

		Convey("When lambda is created", func() {

			Convey("It should provide proper infrastructure", nil)

		})

		Convey("When service binding is created", func() {

			Convey("It should be containing information on lambda's triggers", nil)

		})

		Convey("When event is sent", func() {

			Convey("Lambda should react to the event", nil)

		})

	})

}

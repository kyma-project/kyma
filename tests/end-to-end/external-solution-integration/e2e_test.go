package goconvey_e2e_example

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

const (
	APPNAME             = "e2e-app"
	ACCESSLABEL         = "a1b2c3"
	NAMESPACE           = "production"
	INTEGRATIONAMESPACE = "kyma-integration"
)

func TestApplicationCRDCreation(t *testing.T) {

	Convey("Given application data", t, func() {

		Convey("When CRD created", func() {

			K8SClient, err := testkit.NewK8sResourcesClient(INTEGRATIONAMESPACE)
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

func TestSecureConnectionAndRegistration(t *testing.T) {

	Convey("Given external solution", t, func() {

		Convey("When token requested", func() {

			tokenRequestClient, err := testkit.NewTokenRequestClient(INTEGRATIONAMESPACE)
			So(err, ShouldBeNil)

			tokenRequest, err := tokenRequestClient.CreateTokenRequest(APPNAME)
			So(err, ShouldBeNil)
			So(tokenRequest, ShouldNotBeNil)

			Convey("The operator should insert token into TokenRequest CRD", func() {
				//TODO: Polling of tokenRequest
				time.Sleep(5 * time.Second)

				tokenRequest, err = tokenRequestClient.GetTokenRequest(APPNAME, v1.GetOptions{})
				So(err, ShouldBeNil)
				So(tokenRequest.Status.Token, ShouldNotBeNil)

				Convey("When one-click-integration initiated", func() {

					mockAppClient, err := testkit.NewMockApplicationClient()
					So(err, ShouldBeNil)

					res, err := mockAppClient.ConnectToKyma(tokenRequest.Status.URL, true)
					So(err, ShouldBeNil)
					So(res, ShouldNotBeNil)

					Convey("The app should be connected to Kyma", func() {

						res, err = mockAppClient.GetConnectionInfo()
						So(err, ShouldBeNil)
						So(res.AppName, ShouldEqual, APPNAME)
						So(res.ClusterDomain, ShouldNotBeNil)
						So(res.EventsURL, ShouldNotBeNil)
						So(res.MetadataURL, ShouldNotBeNil)

						Convey("The services and events should be registered", func() {
							apis, err := mockAppClient.GetAPIs()
							So(err, ShouldBeNil)
							So(len(*apis), ShouldBeGreaterThan, 0)
							//TODO: Check whether events/services are properly registered
						})
					})
				})
			})
		})
	})
}

func TestBindings(t *testing.T) {

	Convey("Given application and environment", t, func() {

		lambdaClient, err := testkit.NewLambdaClient(NAMESPACE)
		So(err, ShouldBeNil)

		Convey("When binding is created", func() {

			_, err := lambdaClient.CreateMapping(APPNAME)
			So(err, ShouldBeNil)

			Convey("It should generate proper service classes", func() {

			})
		})
	})
}

func TestLambda(t *testing.T) {

	Convey("Given Lambda's code", t, func() {

		Convey("When lambda and bindings are created", func() {

			lambdaClient, err := testkit.NewLambdaClient(NAMESPACE)
			So(err, ShouldBeNil)

			err = lambdaClient.DeployLambda(APPNAME)
			So(err, ShouldBeNil)

			_, err = lambdaClient.CreateEventActivation(APPNAME)
			So(err, ShouldBeNil)

			_, err = lambdaClient.CreateSubscription(APPNAME)
			So(err, ShouldBeNil)

			//TODO: Create servicebinding for lamda-OCC

			Convey("When event is sent", func() {

				Convey("Lambda should react to the event", func() {

				})
			})
		})
	})
}

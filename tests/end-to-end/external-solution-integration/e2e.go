package main

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testsuite"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"time"
)

func main() {
	time.Sleep(60 * time.Second)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	ts, err := testsuite.NewTestSuite(config, logrus.New())
	if err != nil {
		panic(err)
	}

	err = ts.CreateResources()
	if err != nil {
		panic(err)
	}

	//cert, err := ts.FetchCertificate()

	id, err := ts.RegisterService()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("ID:", id)

	err = ts.CreateInstance(id)
	if err != nil {
		panic(err)
	}
}

//TODO: WIP, delete everything and provide integration with upgrade and backup interfaces along with 'basic' scenario
//import (
//	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
//	. "github.com/smartystreets/goconvey/convey"
//	"k8s.io/apimachinery/pkg/apis/meta/v1"
//	"testing"
//	"time"
//)
//
//const (
//	appName              = "e2e-app"
//	accessLabel          = "a1b2c3"
//	namespace            = "production"
//	integrationNamespace = "kyma-integration"
//)
//
//func TestApplicationCRDCreation(t *testing.T) {
//
//	Convey("Given application data", t, func() {
//
//		Convey("When CRD created", func() {
//
//			K8SClient, err := testkit.NewK8sResourcesClient(integrationNamespace)
//			So(err, ShouldBeNil)
//
//			appConnectorClient, err := testkit.NewAppConnectorClient(integrationNamespace)
//			So(err, ShouldBeNil)
//
//			app, err := appConnectorClient.CreateDummyApplication(appName, accessLabel, false)
//			So(err, ShouldBeNil)
//			So(app, ShouldNotBeNil)
//
//			Convey("The data within CRD should match Application Data", func() {
//
//				app, err := appConnectorClient.GetApplication(appName, v1.GetOptions{})
//				So(err, ShouldBeNil)
//				So(app.Name, ShouldEqual, appName)
//				So(app.Spec.AccessLabel, ShouldEqual, accessLabel)
//				So(app.Spec.SkipInstallation, ShouldEqual, false)
//
//				time.Sleep(5 * time.Second)
//
//				checker := testkit.NewK8sChecker(K8SClient, appName)
//				err = checker.CheckK8sResources()
//
//				So(err, ShouldBeNil)
//			})
//		})
//	})
//}
//
//func TestSecureConnectionAndRegistration(t *testing.T) {
//
//	Convey("Given external solution", t, func() {
//
//		Convey("When token requested", func() {
//
//			tokenRequestClient, err := testkit.NewAppConnectorClient(integrationNamespace)
//			So(err, ShouldBeNil)
//
//			tokenRequest, err := tokenRequestClient.CreateTokenRequest(appName)
//			So(err, ShouldBeNil)
//			So(tokenRequest, ShouldNotBeNil)
//
//			Convey("The operator should insert token into TokenRequest CRD", func() {
//				//TODO: Polling of tokenRequest (may not be needed, needs double-check)
//				time.Sleep(5 * time.Second)
//
//				tokenRequest, err = tokenRequestClient.GetTokenRequest(appName, v1.GetOptions{})
//				So(err, ShouldBeNil)
//				So(tokenRequest.Status.Token, ShouldNotBeNil)
//
//				Convey("When one-click-integration initiated", func() {
//
//					mockAppClient, err := testkit.NewMockApplicationClient()
//					So(err, ShouldBeNil)
//
//					res, err := mockAppClient.ConnectToKyma(tokenRequest.Status.URL, true)
//					So(err, ShouldBeNil)
//					So(res, ShouldNotBeNil)
//
//					Convey("The app should be connected to Kyma", func() {
//
//						res, err = mockAppClient.GetConnectionInfo()
//						So(err, ShouldBeNil)
//						So(res.AppName, ShouldEqual, appName)
//						So(res.ClusterDomain, ShouldNotBeNil)
//						So(res.EventsURL, ShouldNotBeNil)
//						So(res.MetadataURL, ShouldNotBeNil)
//
//						Convey("The services and events should be registered", func() {
//							apis, err := mockAppClient.GetAPIs()
//							So(err, ShouldBeNil)
//							So(len(*apis), ShouldBeGreaterThan, 0)
//							//TODO: Check whether events/services are properly registered
//						})
//					})
//				})
//			})
//		})
//	})
//}
//
//func TestBindings(t *testing.T) {
//
//	Convey("Given application and environment", t, func() {
//
//		eventingClient, err := testkit.NewEventingClient(namespace)
//		So(err, ShouldBeNil)
//
//		Convey("When mapping is created", func() {
//
//			_, err := eventingClient.CreateMapping(appName)
//			So(err, ShouldBeNil)
//
//			Convey("It should generate proper service classes", func() {
//				//TODO: Check service classes
//			})
//		})
//	})
//}
//
//func TestLambda(t *testing.T) {
//
//	Convey("Given Lambda's code", t, func() {
//
//		Convey("When lambda and bindings are created", func() {
//
//			lambdaClient, err := testkit.NewLambdaClient(namespace)
//			So(err, ShouldBeNil)
//
//			err = lambdaClient.DeployLambda(appName)
//			So(err, ShouldBeNil)
//
//			eventingClient, err := testkit.NewEventingClient(namespace)
//			So(err, ShouldBeNil)
//
//			_, err = eventingClient.CreateEventActivation(appName)
//			So(err, ShouldBeNil)
//
//			_, err = eventingClient.CreateSubscription(appName)
//			So(err, ShouldBeNil)
//
//			//TODO: Create servicebinding for lamda-OCC
//
//			Convey("When event is sent", func() {
//
//				//TODO: Send an event via mock-app
//
//				Convey("Lambda should react to the event", func() {
//
//					//TODO: Check response / mock-app / database state
//
//				})
//			})
//		})
//	})
//}
//
//func TestCleanup(t *testing.T) {
//	Convey("Given delete options", t, func() {
//
//		deleteOptions := &v1.DeleteOptions{}
//
//		Convey("It should delete a Lambda", func() {
//
//			lambdaClient, err := testkit.NewLambdaClient(namespace)
//			So(err, ShouldBeNil)
//
//			err = lambdaClient.DeleteLambda(appName, deleteOptions)
//			So(err, ShouldBeNil)
//		})
//
//		Convey("It should delete Subscriptions, EventActivations and ApplicationMappings", func() {
//
//			eventingClient, err := testkit.NewEventingClient(namespace)
//			So(err, ShouldBeNil)
//
//			err = eventingClient.DeleteMapping(appName, deleteOptions)
//			So(err, ShouldBeNil)
//
//			err = eventingClient.DeleteEventActivation(appName, deleteOptions)
//			So(err, ShouldBeNil)
//
//			err = eventingClient.DeleteSubscription(appName, deleteOptions)
//			So(err, ShouldBeNil)
//		})
//
//		Convey("It should delete TokenRequest", func() {
//
//			tokenRequestClient, err := testkit.NewAppConnectorClient(integrationNamespace)
//			So(err, ShouldBeNil)
//
//			err = tokenRequestClient.DeleteTokenRequest(appName, deleteOptions)
//			So(err, ShouldBeNil)
//		})
//
//		Convey("It should delete Application", func() {
//
//			appConnectorClient, err := testkit.NewAppConnectorClient(integrationNamespace)
//			So(err, ShouldBeNil)
//
//			err = appConnectorClient.DeleteApplication(appName, deleteOptions)
//			So(err, ShouldBeNil)
//		})
//
//		//TODO: Delete servicebinding
//	})
//}

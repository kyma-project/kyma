package testsuite

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/resourceskit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"time"
)

type TestSuite interface {
	CreateResources() error
	FetchCertificate() ([]*x509.Certificate, error)
	RegisterService() (string, error)
	CreateInstance(serviceID string) (*v1beta1.ServiceInstance, error)
	StartTestServer() error
	SendEvent()
	CleanUp() error
}

type testSuite struct {
	acClient       resourceskit.AppConnectorClient
	k8sClient      resourceskit.K8sResourcesClient
	trClient       resourceskit.TokenRequestClient
	eventingClient resourceskit.EventingClient
	lambdaClient   resourceskit.LambdaClient
	scClient       resourceskit.ServiceCatalogClient
	testService    testkit.TestService
	connClient     testkit.ConnectorClient
	registryClient testkit.RegistryClient
}

func NewTestSuite(config *rest.Config, logger logrus.FieldLogger) (TestSuite, error) {
	acClient, err := resourceskit.NewAppConnectorClient(config, integrationNamespace)
	if err != nil {
		return nil, err
	}

	k8sClient, err := resourceskit.NewK8sResourcesClient(config, integrationNamespace)
	if err != nil {
		return nil, err
	}

	trClient, err := resourceskit.NewTokenRequestClient(config, integrationNamespace)
	if err != nil {
		return nil, err
	}

	eventingClient, err := resourceskit.NewEventingClient(config, productionNamespace)
	if err != nil {
		return nil, err
	}

	lambdaClient, err := resourceskit.NewLambdaClient(config, productionNamespace)
	if err != nil {
		return nil, err
	}

	scClient, err := resourceskit.NewServiceCatalogClient(config, productionNamespace)
	if err != nil {
		return nil, err
	}

	testService := testkit.NewTestService(k8sClient)

	connClient := testkit.NewConnectorClient(trClient, true, logger)
	registryClient := testkit.NewRegistryClient("http://application-registry-external-api.kyma-integration.svc.cluster.local:8081/"+appName+"/v1/metadata/services", true, logger)

	return &testSuite{
		acClient:       acClient,
		k8sClient:      k8sClient,
		trClient:       trClient,
		connClient:     connClient,
		eventingClient: eventingClient,
		lambdaClient:   lambdaClient,
		registryClient: registryClient,
		scClient:       scClient,
		testService:    testService,
	}, nil
}

func (ts *testSuite) CreateResources() error {
	err := ts.createApplication()
	if err != nil {
		return err
	}

	_, err = ts.eventingClient.CreateMapping(appName)
	if err != nil {
		return err
	}

	_, err = ts.eventingClient.CreateEventActivation(appName)
	if err != nil {
		return err
	}

	//TODO: Get URL from test service and pass it to the lambda
	err = ts.lambdaClient.DeployLambda(appName)
	if err != nil {
		return err
	}

	_, err = ts.eventingClient.CreateSubscription(appName, lambdaEndpoint, eventType, eventVersion)
	if err != nil {
		return err
	}

	return nil
}

func (ts *testSuite) createApplication() error {
	_, err := ts.acClient.CreateDummyApplication(appName, accessLabel, false)
	if err != nil {
		return err
	}

	err = waitUntil(5, 15, ts.isApplicationReady)

	if err != nil {
		return err
	}

	checker := resourceskit.NewK8sChecker(ts.k8sClient, appName)

	err = checker.CheckK8sResources()
	if err != nil {
		return err
	}

	return nil
}

func (ts *testSuite) isApplicationReady() (bool, error) {
	application, e := ts.acClient.GetApplication(appName, v1.GetOptions{})

	if e != nil {
		return false, e
	}

	return application.Status.InstallationStatus.Status == "DEPLOYED", nil
}

func (ts *testSuite) FetchCertificate() ([]*x509.Certificate, error) {
	key, err := testkit.CreateKey()
	if err != nil {
		return nil, err
	}

	infoURL, err := ts.connClient.GetToken(appName)
	if err != nil {
		return nil, err
	}

	info, err := ts.connClient.GetInfo(infoURL)
	if err != nil {
		return nil, err
	}

	csr, err := testkit.CreateCSR(info.Certificate.Subject, key)
	if err != nil {
		return nil, err
	}

	certificate, err := ts.connClient.GetCertificate(info.CertUrl, csr)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}

func (ts *testSuite) RegisterService() (string, error) {
	service := prepareService()

	return ts.registryClient.RegisterService(service)
}

func prepareService() *testkit.ServiceDetails {
	return &testkit.ServiceDetails{
		Provider:         serviceProvider,
		Name:             serviceName,
		Description:      serviceDescription,
		ShortDescription: serviceShortDescription,
		Identifier:       serviceIdentifier,
		Events: &testkit.Events{
			Spec: json.RawMessage(serviceEventsSpec),
		},
	}
}

func (ts *testSuite) CreateInstance(serviceID string) (*v1beta1.ServiceInstance, error) {
	return ts.scClient.CreateServiceInstance(serviceInstanceName, serviceInstanceID, serviceID)
}

func (ts *testSuite) SendEvent() {
	event := prepareEvent()

	testkit.SendEvent("http://application-registry-external-api.kyma-integration.svc.cluster.local:8081/"+appName+"/v1/events", event)
}

func prepareEvent() *testkit.ExampleEvent {
	return &testkit.ExampleEvent{
		EventType:        eventType,
		EventTypeVersion: eventVersion,
		EventID:          "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		EventTime:        time.Now(),
		Data:             "some data",
	}
}

func (ts *testSuite) StartTestServer() error {
	var e error
	e = ts.testService.CreateTestService()

	if e != nil {
		return e
	}

	e = waitUntil(5, 30, ts.testService.CheckIfReady)

	if e != nil {
		return errors.New(fmt.Sprintf("Test Service not started: %s", e))
	}

	return nil
}

func (ts *testSuite) CleanUp() error {
	var e error

	e = ts.testService.DeleteTestService()
	if e != nil {
		return e
	}
	e = ts.lambdaClient.DeleteLambda(appName, &v1.DeleteOptions{})
	if e != nil {
		return e
	}
	e = ts.eventingClient.DeleteSubscription(appName, &v1.DeleteOptions{})
	if e != nil {
		return e
	}
	e = ts.eventingClient.DeleteEventActivation(appName, &v1.DeleteOptions{})
	if e != nil {
		return e
	}
	e = ts.eventingClient.DeleteMapping(appName, &v1.DeleteOptions{})
	if e != nil {
		return e
	}
	e = ts.acClient.DeleteApplication(appName, &v1.DeleteOptions{})
	if e != nil {
		return e
	}
	return nil
}

func waitUntil(retries int, sleepTimeSeconds int, predicate func() (bool, error)) error {
	var ready bool
	var e error

	sleepDuration := time.Duration(sleepTimeSeconds) * time.Second

	for i := 0; i < retries && !ready; i++ {
		ready, e = predicate()
		if e != nil {
			return e
		}
		time.Sleep(sleepDuration)
	}

	if ready {
		return nil
	}

	return errors.New("resource not ready")
}

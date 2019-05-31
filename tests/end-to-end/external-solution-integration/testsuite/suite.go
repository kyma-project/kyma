package testsuite

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/resourceskit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/wait"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
)

type TestSuite interface {
	CreateResources() error
	FetchCertificate() ([]*x509.Certificate, error)
	RegisterService(targetURL string) (string, error)
	CreateInstance(serviceID string) (*v1beta1.ServiceInstance, error)
	CreateServiceBinding() error
	CreateServiceBindingUsage() error
	StartTestServer() error
	SendEvent() error
	CleanUp() error
	GetTestServiceURL() string
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
	serviceID      string
}

func NewTestSuite(config *rest.Config, logger log.FieldLogger) (TestSuite, error) {
	acClient, err := resourceskit.NewAppConnectorClient(config, consts.IntegrationNamespace)
	if err != nil {
		return nil, err
	}

	k8sClient, err := resourceskit.NewK8sResourcesClient(config, consts.IntegrationNamespace)
	if err != nil {
		return nil, err
	}

	trClient, err := resourceskit.NewTokenRequestClient(config, consts.IntegrationNamespace)
	if err != nil {
		return nil, err
	}

	eventingClient, err := resourceskit.NewEventingClient(config, consts.ProductionNamespace)
	if err != nil {
		return nil, err
	}

	lambdaClient, err := resourceskit.NewLambdaClient(config, consts.ProductionNamespace)
	if err != nil {
		return nil, err
	}

	scClient, err := resourceskit.NewServiceCatalogClient(config, consts.ProductionNamespace)
	if err != nil {
		return nil, err
	}

	testService := testkit.NewTestService(k8sClient)

	connClient := testkit.NewConnectorClient(trClient, true, logger)
	registryClient := testkit.NewRegistryClient("http://application-registry-external-api.kyma-integration.svc.cluster.local:8081/"+consts.AppName+"/v1/metadata/services", true, logger)

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

	_, err = ts.eventingClient.CreateMapping()
	if err != nil {
		return err
	}

	_, err = ts.eventingClient.CreateEventActivation()
	if err != nil {
		return err
	}

	err = ts.lambdaClient.DeployLambda()
	if err != nil {
		return err
	}

	err = wait.Until(5, 30, ts.lambdaClient.IsLambdaReady)

	if err != nil {
		return fmt.Errorf("Lambda function not ready, %s", err)
	}

	_, err = ts.eventingClient.CreateSubscription()
	if err != nil {
		return err
	}

	return nil
}

func (ts *testSuite) GetTestServiceURL() string {
	return ts.testService.GetTestServiceURL()
}

func (ts *testSuite) createApplication() error {
	_, err := ts.acClient.CreateDummyApplication(false)
	if err != nil {
		log.Error(err)
		return err
	}
	err = wait.Until(5, 30, ts.isApplicationReady)
	if err != nil {
		log.Error(err)
		return err
	}
	//TODO: Enable this
	// checker := resourceskit.NewK8sChecker(ts.k8sClient)

	// err = checker.CheckK8sResources()
	// if err != nil {
	// 	log.Error(err)
	// 	return err
	// }

	return nil
}

func (ts *testSuite) isApplicationReady() (bool, error) {
	application, err := ts.acClient.GetApplication()

	if err != nil {
		return false, err
	}

	log.Debug("Application installation:", application.Status)
	return application.Status.InstallationStatus.Status == "DEPLOYED", nil
}

func (ts *testSuite) FetchCertificate() ([]*x509.Certificate, error) {
	key, err := testkit.CreateKey()
	if err != nil {
		return nil, err
	}

	infoURL, err := ts.connClient.GetToken()
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

func (ts *testSuite) RegisterService(targetURL string) (string, error) {
	service := prepareService(targetURL)

	id, err := ts.registryClient.RegisterService(service)
	if err == nil {
		ts.serviceID = id
	}
	return id, err
}

func prepareService(targetURL string) *testkit.ServiceDetails {
	return &testkit.ServiceDetails{
		Provider:         consts.ServiceProvider,
		Name:             consts.ServiceName,
		Description:      consts.ServiceDescription,
		ShortDescription: consts.ServiceShortDescription,
		Identifier:       consts.ServiceIdentifier,
		Events: &testkit.Events{
			Spec: json.RawMessage(consts.ServiceEventsSpec),
		},
		Api: &testkit.API{
			TargetUrl: targetURL,
		},
	}
}

func (ts *testSuite) CreateInstance(serviceID string) (*v1beta1.ServiceInstance, error) {
	return ts.scClient.CreateServiceInstance(consts.ServiceInstanceName, consts.ServiceInstanceID, serviceID)
}

func (ts *testSuite) CreateServiceBinding() error {
	_, err := ts.scClient.CreateServiceBinding()
	if err != nil {
		return err
	}

	return nil
}

func (ts *testSuite) CreateServiceBindingUsage() error {
	_, err := ts.scClient.CreateServiceBindingUsage()
	if err != nil {
		return err
	}

	err = wait.Until(5, 30, ts.lambdaClient.IsFunctionAnnotated)
	if err != nil {
		return err
	}

	err = wait.Until(5, 30, ts.lambdaClient.IsLambdaReadyWithSBU)
	if err != nil {
		return err
	}
	return nil
}

func (ts *testSuite) SendEvent() error {
	event := prepareEvent()

	return testkit.SendEvent("http://application-registry-external-api.kyma-integration.svc.cluster.local:8081/"+consts.AppName+"/v1/events", event)
}

func prepareEvent() *testkit.ExampleEvent {
	return &testkit.ExampleEvent{
		EventType:        consts.EventType,
		EventTypeVersion: consts.EventVersion,
		EventID:          "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		EventTime:        time.Now(),
		Data:             "some data",
	}
}

func (ts *testSuite) StartTestServer() error {
	err := ts.testService.CreateTestService()

	if err != nil {
		return err
	}

	err = wait.Until(5, 30, ts.testService.IsReady)

	if err != nil {
		return fmt.Errorf("Test Service not started: %s", err)
	}

	return nil
}

func (ts *testSuite) CleanUp() error {
	return nil
	errorOccured := false
	var err error

	err = ts.lambdaClient.DeleteLambda()
	if err != nil {
		log.Error(err)
		errorOccured = true
	}

	err = ts.scClient.DeleteServiceBindingUsage()
	if err != nil {
		log.Error(err)
		errorOccured = true
	}

	err = ts.scClient.DeleteServiceBinding()
	if err != nil {
		log.Error(err)
		errorOccured = true
	}

	err = ts.scClient.DeleteServiceInstance(consts.ServiceInstanceName)
	if err != nil {
		log.Error(err)
		errorOccured = true
	}

	err = ts.testService.DeleteTestService()
	if err != nil {
		log.Error(err)
		errorOccured = true
	}

	err = ts.eventingClient.DeleteSubscription()
	if err != nil {
		log.Error(err)
		errorOccured = true
	}

	err = ts.eventingClient.DeleteEventActivation()
	if err != nil {
		log.Error(err)
		errorOccured = true
	}

	err = ts.eventingClient.DeleteMapping()
	if err != nil {
		log.Error(err)
		errorOccured = true
	}

	err = ts.acClient.DeleteApplication()
	if err != nil {
		log.Error(err)
		errorOccured = true
	}

	if ts.serviceID != "" {
		err = ts.registryClient.DeleteService(ts.serviceID)
		if err != nil {
			log.Error(err)
			errorOccured = true
		}
	}

	if errorOccured {
		return fmt.Errorf("cleanup failed")
	}
	return nil
}

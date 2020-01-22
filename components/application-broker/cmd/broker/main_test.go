package main

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	istioversionedclientfake "istio.io/client-go/pkg/clientset/versioned/fake"

	scfake "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	bt "github.com/kyma-project/kyma/components/application-broker/internal/broker/testing"
	"github.com/kyma-project/kyma/components/application-broker/internal/config"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	abfake "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	abCS "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
	v1alpha12 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appfake "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/fake"
	appCS "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

const (
	namespace       = "production"
	applicationName = "ec-prod"

	serviceOneID = "001"
	serviceTwoID = "002"
)

type testSuite struct {
	t         *testing.T
	server    *httptest.Server
	osbClient osb.Client

	abInterface  abCS.ApplicationconnectorV1alpha1Interface
	appInterface appCS.ApplicationconnectorV1alpha1Interface
	scInterface  sc.ServicecatalogV1beta1Interface

	serviceID string
	planID    string
}

func TestGetCatalogHappyPath(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()

	suite.createApplication()

	// when
	suite.enableApplication()

	// then
	suite.AssertServicesInCatalogEndpoint(serviceOneID, serviceTwoID)
	suite.AssertServiceBrokerRegistered()
}

func TestRegisterAndUnregisterServiceBroker(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()

	suite.createApplication()

	// when
	suite.enableApplication()

	// then
	suite.AssertServicesInCatalogEndpoint(serviceOneID, serviceTwoID)
	suite.AssertServiceBrokerRegistered()

	// when
	suite.disableApplication()

	//then
	suite.AssertServiceBrokerNotRegistered()
}

func TestUnregistrationServiceBrokerBlockedByExistingInstance(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()

	suite.createApplication()
	suite.enableApplication()
	suite.AssertServicesInCatalogEndpoint(serviceOneID, serviceTwoID)
	suite.AssertServiceBrokerRegistered()
	suite.ProvisionInstance("instance-01")

	// when
	suite.disableApplication()

	// assert the offering is empty
	suite.AssertServicesInCatalogEndpoint()

	// then
	// ServiceBroker still exists because of existing instance
	suite.AssertServiceBrokerRegistered()

	// when
	suite.DeprovisionInstance("instance-01")

	// then
	suite.AssertServiceBrokerNotRegistered()
}

func TestGetCatalogEnableSelectedServices(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()

	suite.createApplication()

	// when
	suite.enableApplicationServices(serviceOneID)

	// then
	suite.AssertServicesInCatalogEndpoint(serviceOneID)
}

func newTestSuite(t *testing.T) *testSuite {
	log := spy.NewLogSink()

	cfg := config.Config{
		Namespace:                  "kyma-system",
		ServiceName:                "application-broker",
		Port:                       8001,
		BrokerRelistDurationWindow: time.Hour,
		Storage: []storage.Config{
			{Driver: storage.DriverMemory, Provide: storage.ProviderConfigMap{storage.EntityAll: storage.ProviderConfig{}}},
		},
	}

	stopCh := make(chan struct{})

	abClientSet := abfake.NewSimpleClientset()
	k8sClientSet := k8sfake.NewSimpleClientset()
	scClientSet := scfake.NewSimpleClientset()
	appClient := appfake.NewSimpleClientset()
	knClient := knative.NewClient(bt.NewFakeClients())
	istioClient := istioversionedclientfake.NewSimpleClientset()

	k8sClientSet.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})

	livenessCheckStatus := broker.LivenessCheckStatus{Succeeded: false}

	srv := SetupServerAndRunControllers(&cfg, log.Logger, stopCh, k8sClientSet, scClientSet, appClient, abClientSet,
		knClient, istioClient, &livenessCheckStatus)
	server := httptest.NewServer(srv.CreateHandler())

	osbClient, err := newOSBClient(fmt.Sprintf("%s/%s", server.URL, namespace))
	require.NoError(t, err)

	return &testSuite{
		server:       server,
		abInterface:  abClientSet.ApplicationconnectorV1alpha1(),
		appInterface: appClient.ApplicationconnectorV1alpha1(),
		osbClient:    osbClient,
		scInterface:  scClientSet.ServicecatalogV1beta1(),
		t:            t,
	}
}

func newOSBClient(url string) (osb.Client, error) {
	config := osb.DefaultClientConfiguration()
	config.URL = url
	config.APIVersion = osb.Version2_13()

	osbClient, err := osb.NewClient(config)
	if err != nil {
		return nil, err
	}

	return osbClient, nil
}

func (ts *testSuite) createApplication() {
	_, err := ts.appInterface.Applications().Create(&v1alpha12.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: applicationName,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha12.SchemeGroupVersion.String(),
		},
		Spec: v1alpha12.ApplicationSpec{
			Description: "EC description",
			AccessLabel: "access-label",
			Services: []v1alpha12.Service{
				{
					ID:                  serviceOneID,
					DisplayName:         "name1",
					Tags:                []string{"tag1", "tag2"},
					LongDescription:     "desc1",
					ProviderDisplayName: "name1",
					Labels:              map[string]string{"connected-app": "ec-prod"},
					Entries: []v1alpha12.Entry{
						{
							Type:        "API",
							GatewayUrl:  "url",
							AccessLabel: "label",
						},
						{
							Type: "Events",
						},
					},
				},
				{
					ID:                  serviceTwoID,
					DisplayName:         "name2",
					Tags:                []string{"tag1", "tag2"},
					LongDescription:     "desc2",
					ProviderDisplayName: "name2",
					Labels:              map[string]string{"connected-app": "ec-prod"},
					Entries: []v1alpha12.Entry{
						{
							Type:        "API",
							GatewayUrl:  "url",
							AccessLabel: "label",
						},
						{
							Type: "Events",
						},
					},
				},
			},
		},
	})
	require.NoError(ts.t, err)
}

func (ts *testSuite) AssertServiceBrokerRegistered() {
	timeoutCh := time.After(3 * time.Second)
	for {
		_, err := ts.scInterface.ServiceBrokers(namespace).Get(nsbroker.NamespacedBrokerName, metav1.GetOptions{})
		if err == nil {
			return
		}
		select {
		case <-timeoutCh:
			assert.Fail(ts.t, "The timeout exceeded while waiting for the ServiceBroker", err)
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (ts *testSuite) AssertServiceBrokerNotRegistered() {
	timeoutCh := time.After(3 * time.Second)
	for {
		_, err := ts.scInterface.ServiceBrokers(namespace).Get(nsbroker.NamespacedBrokerName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return
		}
		select {
		case <-timeoutCh:
			assert.Fail(ts.t, "The timeout exceeded while waiting for the ServiceBroker unregistration")
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

}

func (ts *testSuite) ProvisionInstance(id string) {
	osbResponse, err := ts.osbClient.GetCatalog()
	require.NoError(ts.t, err)
	svcID := osbResponse.Services[0].ID
	planID := osbResponse.Services[0].Plans[0].ID
	_, err = ts.osbClient.ProvisionInstance(&osb.ProvisionRequest{
		PlanID:            planID,
		ServiceID:         svcID,
		InstanceID:        id,
		OrganizationGUID:  "org-guid",
		SpaceGUID:         "spaceGUID",
		AcceptsIncomplete: true,
	})
	require.NoError(ts.t, err)

	// save IDs for deprovisioning
	ts.serviceID = svcID
	ts.planID = planID

	// The controller checks if there is any Service Instance (managed by Service Catalog).
	// The following code simulates Service Catalog actions
	ts.scInterface.ServiceClasses(namespace).Create(&v1beta1.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-class",
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceClassSpec{
			ServiceBrokerName: nsbroker.NamespacedBrokerName,
		},
	})
	ts.scInterface.ServiceInstances(namespace).Create(&v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance-001",
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			ServiceClassRef: &v1beta1.LocalObjectReference{
				Name: "app-class",
			},
		},
	})
}

func (ts *testSuite) DeprovisionInstance(id string) {
	_, err := ts.osbClient.DeprovisionInstance(&osb.DeprovisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        id,
		ServiceID:         ts.serviceID,
		PlanID:            ts.planID,
	})
	require.NoError(ts.t, err)

	// The controller checks if there is any Service Instance (managed by Service Catalog).
	// The following code simulates Service Catalog actions
	ts.scInterface.ServiceInstances(namespace).Delete("instance-001", &metav1.DeleteOptions{})
}

func (ts *testSuite) AssertServicesInCatalogEndpoint(ids ...string) {
	timeoutCh := time.After(3 * time.Second)
	for {
		err := ts.checkServiceIDs(ids...)
		if err == nil {
			return
		}
		select {
		case <-timeoutCh:
			assert.Fail(ts.t, "The timeout exceeded while waiting for the OSB catalog response, last error: %v", err)
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (ts *testSuite) checkServiceIDs(ids ...string) error {
	osbResponse, err := ts.osbClient.GetCatalog()
	if err != nil {
		return err
	}

	services := make(map[string]osb.Service)
	for _, svc := range osbResponse.Services {
		services[svc.ID] = svc
	}
	for _, id := range ids {
		if _, exists := services[id]; !exists {
			return fmt.Errorf("the /v2/catalog response does not contain service with id %s", id)
		}
	}

	for _, svc := range osbResponse.Services {
		found := false
		for _, id := range ids {
			if svc.ID == id {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("the /v2/catalog contains service %s which is not expected", svc.ID)
		}
	}

	return nil
}

func (ts *testSuite) enableApplication() {
	_, err := ts.abInterface.ApplicationMappings(namespace).Create(&v1alpha1.ApplicationMapping{
		ObjectMeta: metav1.ObjectMeta{
			Name:      applicationName,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
	})
	require.NoError(ts.t, err)
}

func (ts *testSuite) disableApplication() {
	err := ts.abInterface.ApplicationMappings(namespace).Delete(applicationName, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)
}

func (ts *testSuite) enableApplicationServices(ids ...string) {
	var services []v1alpha1.ApplicationMappingService
	for _, id := range ids {
		services = append(services, v1alpha1.ApplicationMappingService{ID: id})
	}

	_, err := ts.abInterface.ApplicationMappings(namespace).Create(&v1alpha1.ApplicationMapping{
		ObjectMeta: metav1.ObjectMeta{
			Name:      applicationName,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.ApplicationMappingSpec{
			Services: services,
		},
	})
	require.NoError(ts.t, err)
}

func (ts *testSuite) tearDown() {
	ts.server.Close()
}

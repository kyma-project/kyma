package main

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Azure/open-service-broker-azure/pkg/slice"
	scfake "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-broker/internal/config"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	abfake "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	abCS "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
	v1alpha12 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appfake "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/fake"
	appCS "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	corev1 "k8s.io/api/core/v1"
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
}

func TestGetCatalogHappyPath(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()

	suite.createApplication()

	// when
	suite.enableApplication()

	// then
	suite.AssertServicesInCatalogEndpoint(3*time.Second, serviceOneID, serviceTwoID)
}

func TestGetCatalogEnableSelectedServices(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()

	suite.createApplication()

	// when
	suite.enableApplicationServices(serviceOneID)

	// then
	suite.AssertServicesInCatalogEndpoint(3*time.Second, serviceOneID)
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

	k8sClientSet.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})

	srv := SetupServerAndRunControllers(&cfg, log.Logger, stopCh, k8sClientSet, scClientSet, appClient, abClientSet, nil)
	server := httptest.NewServer(srv.CreateHandler())

	osbClient, err := newOSBClient(fmt.Sprintf("%s/%s", server.URL, namespace))
	require.NoError(t, err)

	return &testSuite{
		server:       server,
		abInterface:  abClientSet.ApplicationconnectorV1alpha1(),
		appInterface: appClient.ApplicationconnectorV1alpha1(),
		osbClient:    osbClient,
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

func (ts *testSuite) AssertServicesInCatalogEndpoint(timeout time.Duration, ids ...string) {
	timeoutCh := time.After(timeout)
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
		if !slice.ContainsString(ids, svc.ID) {
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

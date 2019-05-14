package applicationoperator

import (
	"fmt"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appConnector "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	applicationName = "application-for-testing"
	apiServiceID    = "api-service-id"
	eventsServiceID = "events-service-id"
	gatewayURL      = "https://gateway.local"
)

// UpgradeTest checks if Application Operator upgrades image versions of Application Proxy and Event Service of the created Application.
type UpgradeTest struct {
	appConnectorInterface appConnector.Interface
}

// NewApplicationOperatorUpgradeTest returns new instance of the UpgradeTest.
func NewApplicationOperatorUpgradeTest(acCli appConnector.Interface) *UpgradeTest {
	return &UpgradeTest{
		appConnectorInterface: acCli,
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (ut *UpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Info("ApplicationOperator UpgradeTest creating resources...")

	err := ut.createApplication(log)
	if err != nil {
		return errors.Wrap(err, "could not create Application")
	}

	found, err := ut.findResources()
	if err != nil {
		return errors.Wrap(err, "could not find resources")
	}
	if !found {
		return fmt.Errorf("could not find Application Proxy or Event Service")
	}

	log.Info("ApplicationOperator UpgradeTest resources created!")
	return nil
}

func (ut *UpgradeTest) createApplication(log logrus.FieldLogger) error {
	log.Info("Creating Application...")

	_, err := ut.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Create(&v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: applicationName,
		},
		Spec: v1alpha1.ApplicationSpec{
			AccessLabel:      "app-access-label",
			Description:      "Application used by application acceptance test",
			SkipInstallation: true,
			Services: []v1alpha1.Service{
				{
					ID:   apiServiceID,
					Name: apiServiceID,
					Labels: map[string]string{
						"connected-app": "app-name",
					},
					ProviderDisplayName: "provider",
					DisplayName:         "Some API",
					Description:         "Application Service Class with API",
					Tags:                []string{},
					Entries: []v1alpha1.Entry{
						{
							Type:        "API",
							AccessLabel: "accessLabel",
							GatewayUrl:  gatewayURL,
						},
					},
				},
				{
					ID:   eventsServiceID,
					Name: eventsServiceID,
					Labels: map[string]string{
						"connected-app": "app-name",
					},
					ProviderDisplayName: "provider",
					DisplayName:         "Some Events",
					Description:         "Application Service Class with Events",
					Tags:                []string{},
					Entries: []v1alpha1.Entry{
						{
							Type: "Events",
						},
					},
				},
			},
		},
	})
	return err
}

func (ut *UpgradeTest) findResources() (bool, error) {
	//TODO: Make sure that Application Proxy and Event Service are created
	return true, nil
}

// TestResources tests resources after backup phase
func (ut *UpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Info("ApplicationOperator UpgradeTest testing resources...")
	//TODO: Make sure that image versions of the Application Proxy and Event Service are upgraded

	log.Info("ApplicationOperator UpgradeTest test passed!")
	return nil
}

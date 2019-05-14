package applicationoperator

import (
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"k8s.io/apimachinery/pkg/util/wait"

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
	appConnectorInterface   appConnector.Interface
	serviceCatalogInterface clientset.Interface
}

// NewApplicationOperatorUpgradeTest returns new instance of the UpgradeTest.
func NewApplicationOperatorUpgradeTest(acCli appConnector.Interface, scCli clientset.Interface) *UpgradeTest {
	return &UpgradeTest{
		appConnectorInterface:   acCli,
		serviceCatalogInterface: scCli,
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (ut *UpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Info("ApplicationOperator UpgradeTest creating resources...")

	if err := ut.createApplication(log); err != nil {
		return errors.Wrap(err, "could not create Application")
	}

	if err := ut.waitForResources(stop, log, namespace); err != nil {
		return errors.Wrap(err, "could not find resources")
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
			SkipInstallation: true, //TODO: Make sure that "true" value is correct here
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

func (ut *UpgradeTest) waitForResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	//TODO: Make sure that Application Proxy is created by Application Operator

	if err := ut.waitForInstance(eventsServiceID, namespace, stop); err != nil {
		return errors.Wrap(err, "could not find Event Service instance")
	}

	return nil
}

func (ut *UpgradeTest) waitForInstance(name, namespace string, stop <-chan struct{}) error {
	return ut.wait(2*time.Minute, func() (done bool, err error) {
		si, err := ut.serviceCatalogInterface.ServicecatalogV1beta1().ServiceInstances(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if si.Status.ProvisionStatus == v1beta1.ServiceInstanceProvisionStatusProvisioned {
			return true, nil
		}
		return false, nil
	}, stop)
}

func (ut *UpgradeTest) wait(timeout time.Duration, conditionFunc wait.ConditionFunc, stop <-chan struct{}) error {
	timeoutCh := time.After(timeout)
	stopCh := make(chan struct{})
	go func() {
		select {
		case <-timeoutCh:
			close(stopCh)
		case <-stop:
			close(stopCh)
		}
	}()
	return wait.PollUntil(500*time.Millisecond, conditionFunc, stopCh)
}

// TestResources tests resources after backup phase
func (ut *UpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Info("ApplicationOperator UpgradeTest testing resources...")
	//TODO: Make sure that image versions of the Application Proxy and Event Service are upgraded

	if err := ut.deleteResources(log, namespace); err != nil {
		return errors.Wrap(err, "could not delete resources")
	}

	log.Info("ApplicationOperator UpgradeTest test passed!")
	return nil
}

func (ut *UpgradeTest) deleteResources(log logrus.FieldLogger, namespace string) error {
	if err := ut.deleteApplication(log); err != nil {
		return err
	}

	if err := ut.deleteServiceInstance(eventsServiceID, namespace); err != nil {
		return err
	}

	//TODO: Delete Application Proxy and other created resources

	return nil
}

func (ut *UpgradeTest) deleteApplication(log logrus.FieldLogger) error {
	return ut.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Delete(applicationName, &metav1.DeleteOptions{})
}

func (ut *UpgradeTest) deleteServiceInstance(name, namespace string) error {
	return ut.serviceCatalogInterface.ServicecatalogV1beta1().ServiceInstances(namespace).Delete(name, &metav1.DeleteOptions{})
}

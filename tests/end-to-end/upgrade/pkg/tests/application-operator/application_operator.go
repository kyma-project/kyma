package applicationoperator

import (
	"fmt"
	"time"

	v1 "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appConnector "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsCli "k8s.io/client-go/kubernetes/typed/apps/v1"
)

const (
	applicationName            = "test-app"
	eventsServiceDeployment    = "test-app-event-service"
	applicationProxyDeployment = "test-app-application-gateway"
	apiServiceID               = "api-service-id"
	eventsServiceID            = "events-service-id"
	gatewayURL                 = "https://gateway.local"
)

// UpgradeTest checks if Application Operator upgrades image versions of Application Proxy and Events Service of the created Application.
type UpgradeTest struct {
	appConnectorInterface appConnector.Interface
	deploymentCli         appsCli.DeploymentsGetter
}

// NewApplicationOperatorUpgradeTest returns new instance of the UpgradeTest.
func NewApplicationOperatorUpgradeTest(acCli appConnector.Interface, deployCli appsCli.DeploymentsGetter) *UpgradeTest {
	return &UpgradeTest{
		appConnectorInterface: acCli,
		deploymentCli:         deployCli,
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

	//TODO: Get Events Service and Application Proxy images and timestamps
	//TODO: Create ConfigMap with Events Service and Application Proxy images and timestamps

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
	if err := ut.waitForDeployment(applicationProxyDeployment, namespace, stop); err != nil {
		return errors.Wrap(err, "could not find Application Proxy deployment")
	}

	if err := ut.waitForDeployment(eventsServiceDeployment, namespace, stop); err != nil {
		return errors.Wrap(err, "could not find Events Service deployment")
	}

	return nil
}

func (ut *UpgradeTest) waitForDeployment(name, namespace string, stop <-chan struct{}) error {
	return ut.wait(2*time.Minute, func() (done bool, err error) {
		si, err := ut.deploymentCli.Deployments(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if si.Status.ReadyReplicas >= 1 {
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

	if err := ut.verifyResources(namespace); err != nil {
		return errors.Wrap(err, "image versions are not upgraded")
	}

	if err := ut.deleteResources(log, namespace); err != nil {
		return errors.Wrap(err, "could not delete resources")
	}

	log.Info("ApplicationOperator UpgradeTest test passed!")
	return nil
}

func (ut *UpgradeTest) verifyResources(namespace string) error {
	if err := ut.verifyDeployment(eventsServiceDeployment, namespace); err != nil {
		return errors.Wrap(err, "Events Service is not upgraded")
	}

	if err := ut.verifyDeployment(applicationProxyDeployment, namespace); err != nil {
		return errors.Wrap(err, "Application Proxy is not upgraded")
	}

	return nil
}

func (ut *UpgradeTest) verifyDeployment(name, namespace string) error {
	deployment, err := ut.getDeployment(name, namespace)
	if err != nil {
		return err
	}

	//TODO: Make sure that image version of the deployment is upgraded (use values from ConfigMap)
	// images should be different OR timestamps should be different
	// this condition covers two cases (1. the same image after upgrade 2. new image after upgrade)
	fmt.Println(deployment.Spec.Template.Spec.Containers[0].Image)
	fmt.Println(deployment.CreationTimestamp)

	return nil
}

func (ut *UpgradeTest) getDeployment(name, namespace string) (*v1.Deployment, error) {
	return ut.deploymentCli.Deployments(namespace).Get(name, metav1.GetOptions{})
}

func (ut *UpgradeTest) deleteResources(log logrus.FieldLogger, namespace string) error {
	if err := ut.deleteApplication(); err != nil {
		return err
	}

	//TODO: Make sure that all created resources are deleted

	return nil
}

func (ut *UpgradeTest) deleteApplication() error {
	return ut.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Delete(applicationName, &metav1.DeleteOptions{})
}

func (ut *UpgradeTest) deleteDeployment(name, namespace string) error {
	return ut.deploymentCli.Deployments(namespace).Delete(name, &metav1.DeleteOptions{})
}

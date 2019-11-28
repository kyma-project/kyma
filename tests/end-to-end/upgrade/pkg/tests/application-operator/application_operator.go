package applicationoperator

import (
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/apps/v1"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appConnector "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sCli "k8s.io/client-go/kubernetes"
)

const (
	applicationName                 = "operator-test-app"
	eventsServiceDeployment         = "operator-test-app-event-service"
	applicationGatewayDeployment    = "operator-test-app-application-gateway"
	connectivityValidatorDeployment = "operator-test-app-connectivity-validator"
	integrationNamespace            = "kyma-integration"

	applicationOperatorStatefulset = "application-operator"

	appGatewayImageOptPrefix            = "--applicationGatewayImage="
	eventServiceImageOptPrefix          = "--eventServiceImage="
	connectivityValidatorImageOptPrefix = "--applicationConnectivityValidatorImage="

	imageKey     = "image"
	timestampKey = "timestamp"
)

type appImagesConfig struct {
	eventServiceImage          string
	appGatewayImage            string
	connectivityValidatorImage string
}

// UpgradeTest checks if Application Operator upgrades image versions of Application Proxy and Events Service of the created Application.
type UpgradeTest struct {
	appConnectorInterface appConnector.Interface
	k8sCli                k8sCli.Clientset
}

// NewApplicationOperatorUpgradeTest returns new instance of the UpgradeTest.
func NewApplicationOperatorUpgradeTest(acCli appConnector.Interface, k8sCli k8sCli.Clientset) *UpgradeTest {
	return &UpgradeTest{
		appConnectorInterface: acCli,
		k8sCli:                k8sCli,
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (ut *UpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Info("ApplicationOperator UpgradeTest creating resources...")

	log.Info("Creating Application...")
	if err := ut.createApplication(); err != nil {
		return errors.Wrap(err, "could not create Application")
	}

	log.Info("Waiting for resources...")
	if err := ut.waitForResources(stop); err != nil {
		return errors.Wrap(err, "could not find resources")
	}

	log.Info("ApplicationOperator UpgradeTest is set and ready!")
	return nil
}

func (ut *UpgradeTest) createApplication() error {
	_, err := ut.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Create(&v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: applicationName,
		},
		Spec: v1alpha1.ApplicationSpec{
			AccessLabel: "app-access-label",
			Description: "Application used by upgradability test",
			Services:    []v1alpha1.Service{},
		},
	})
	return err
}

func (ut *UpgradeTest) waitForResources(stop <-chan struct{}) error {
	return ut.wait(2*time.Minute, ut.isApplicationReady, stop)
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

func (ut *UpgradeTest) isApplicationReady() (bool, error) {
	application, err := ut.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Get(applicationName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	return application.Status.InstallationStatus.Status == "DEPLOYED", nil
}

func (ut *UpgradeTest) getImagesConfigFromOperatorOpts() (appImagesConfig, error) {
	statefulSet, err := ut.k8sCli.AppsV1().StatefulSets(integrationNamespace).Get(applicationOperatorStatefulset, metav1.GetOptions{})
	if err != nil {
		return appImagesConfig{}, errors.Wrap(err, "Failed to get Application Operator stateful set")
	}

	containerArgs := statefulSet.Spec.Template.Spec.Containers[0].Args

	appImagesConfig := appImagesConfig{}

	for _, arg := range containerArgs {
		if strings.HasPrefix(arg, appGatewayImageOptPrefix) {
			appImagesConfig.appGatewayImage = strings.TrimPrefix(arg, appGatewayImageOptPrefix)
		}
		if strings.HasPrefix(arg, eventServiceImageOptPrefix) {
			appImagesConfig.eventServiceImage = strings.TrimPrefix(arg, eventServiceImageOptPrefix)
		}
		if strings.HasPrefix(arg, connectivityValidatorImageOptPrefix) {
			appImagesConfig.connectivityValidatorImage = strings.TrimPrefix(arg, connectivityValidatorImageOptPrefix)
		}
	}

	return appImagesConfig, nil
}

func (ut *UpgradeTest) getDeploymentData(name, namespace string) (image string, err error) {
	deployment, err := ut.getDeployment(name, namespace)
	if err != nil {
		return "", err
	}

	imageVersion, err := getImageVersion(name, deployment.Spec.Template.Spec.Containers)
	if err != nil {
		return "", err
	}

	return imageVersion, nil
}

func (ut *UpgradeTest) getDeployment(name, namespace string) (*v1.Deployment, error) {
	return ut.k8sCli.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
}

func getImageVersion(containerName string, containers []core.Container) (string, error) {
	for _, c := range containers {
		if strings.Contains(c.Name, containerName) {
			return c.Image, nil
		}
	}
	return "", fmt.Errorf(fmt.Sprintf("container name %s not found", containerName))
}

// TestResources tests resources after backup phase
func (ut *UpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Info("ApplicationOperator UpgradeTest testing resources...")

	log.Info("Verifying resources...")
	if err := ut.verifyResources(integrationNamespace); err != nil {
		return errors.Wrap(err, "image versions are not upgraded")
	}

	log.Info("Deleting resources...")
	if err := ut.deleteResources(integrationNamespace); err != nil {
		return errors.Wrap(err, "could not delete resources")
	}

	log.Info("ApplicationOperator UpgradeTest test passed!")
	return nil
}

func (ut *UpgradeTest) verifyResources(namespace string) error {
	appImages, err := ut.getImagesConfigFromOperatorOpts()
	if err != nil {
		return errors.Wrap(err, "Failed to get expected images from Stateful set")
	}

	if err := ut.verifyDeployment(eventsServiceDeployment, namespace, appImages.eventServiceImage); err != nil {
		return errors.Wrap(err, "Events Service is not upgraded")
	}

	if err := ut.verifyDeployment(applicationGatewayDeployment, namespace, appImages.appGatewayImage); err != nil {
		return errors.Wrap(err, "Application Gateway is not upgraded")
	}

	if err := ut.verifyDeployment(connectivityValidatorDeployment, namespace, appImages.connectivityValidatorImage); err != nil {
		return errors.Wrap(err, "Application Connectivity Validator is not upgraded")
	}

	return nil
}

func (ut *UpgradeTest) verifyDeployment(name, namespace, expectedImage string) error {
	image, err := ut.getDeploymentData(name, namespace)
	if err != nil {
		return err
	}

	if image != expectedImage {
		return fmt.Errorf("invalid image of %s. Expected: %s. Actual: %s", name, expectedImage, image)
	}

	return nil
}

func (ut *UpgradeTest) getConfigMap(name, namespace string) (*core.ConfigMap, error) {
	return ut.k8sCli.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
}

func (ut *UpgradeTest) deleteResources(namespace string) error {
	if err := ut.deleteApplication(); err != nil {
		return err
	}

	return nil
}

func (ut *UpgradeTest) deleteApplication() error {
	return ut.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Delete(applicationName, &metav1.DeleteOptions{})
}

func (ut *UpgradeTest) deleteConfigMap(name, namespace string) error {
	return ut.k8sCli.CoreV1().ConfigMaps(namespace).Delete(name, &metav1.DeleteOptions{})
}

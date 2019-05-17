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
	applicationName            = "operator-test-app"
	eventsServiceDeployment    = "operator-test-app-event-service"
	applicationProxyDeployment = "operator-test-app-application-gateway"
	eventsConfigMapName        = "operator-test-app-configmap-events"
	proxyConfigMapName         = "operator-test-app-configmap-proxy"
	integrationNamespace       = "kyma-integration"

	imageKey     = "image"
	timestampKey = "timestamp"
)

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

	log.Info("Creating verification data...")
	if err := ut.setVerificationData(integrationNamespace); err != nil {
		return errors.Wrap(err, "could not set verification data")
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

func (ut *UpgradeTest) setVerificationData(namespace string) error {
	eventsImage, eventsTimestamp, err := ut.getDeploymentData(eventsServiceDeployment, namespace)
	if err != nil {
		return err
	}

	if err := ut.createConfigMap(eventsConfigMapName, namespace, eventsImage, eventsTimestamp); err != nil {
		return err
	}

	proxyImage, proxyTimestamp, err := ut.getDeploymentData(applicationProxyDeployment, namespace)
	if err != nil {
		return err
	}

	err = ut.createConfigMap(proxyConfigMapName, namespace, proxyImage, proxyTimestamp)
	return err
}

func (ut *UpgradeTest) getDeploymentData(name, namespace string) (image string, timestamp metav1.Time, err error) {
	deployment, err := ut.getDeployment(name, namespace)
	if err != nil {
		return "", metav1.Time{}, err
	}

	imageVersion, err := getImageVersion(name, deployment.Spec.Template.Spec.Containers)
	if err != nil {
		return "", metav1.Time{}, err
	}

	return imageVersion, deployment.CreationTimestamp, nil
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

func (ut *UpgradeTest) createConfigMap(name, namespace, image string, timestamp metav1.Time) error {
	configMap := &core.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{
			imageKey:     image,
			timestampKey: timestamp.String(),
		},
	}

	_, err := ut.k8sCli.CoreV1().ConfigMaps(namespace).Create(configMap)
	if err != nil {
		return err
	}

	return nil
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
	if err := ut.verifyDeployment(eventsServiceDeployment, eventsConfigMapName, namespace); err != nil {
		return errors.Wrap(err, "Events Service is not upgraded")
	}

	if err := ut.verifyDeployment(applicationProxyDeployment, proxyConfigMapName, namespace); err != nil {
		return errors.Wrap(err, "Application Proxy is not upgraded")
	}

	return nil
}

func (ut *UpgradeTest) verifyDeployment(name, configmapName, namespace string) error {
	image, timestamp, err := ut.getDeploymentData(name, namespace)
	if err != nil {
		return err
	}

	configMap, err := ut.getConfigMap(configmapName, namespace)
	if err != nil {
		return err
	}

	previousImage, ok := configMap.Data[imageKey]
	if !ok {
		return fmt.Errorf("pre-upgrade image not found")
	}

	previousTimestamp, ok := configMap.Data[timestampKey]
	if !ok {
		return fmt.Errorf("pre-upgrade timestamp not found")
	}

	if previousImage == image && previousTimestamp == timestamp.String() {
		return fmt.Errorf("image and timestamp not changed")
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

	if err := ut.deleteConfigMap(proxyConfigMapName, namespace); err != nil {
		return errors.Wrap(err, "Proxy ConfigMap could not be deleted")
	}

	if err := ut.deleteConfigMap(eventsConfigMapName, namespace); err != nil {
		return errors.Wrap(err, "Events ConfigMap could not be deleted")
	}

	return nil
}

func (ut *UpgradeTest) deleteApplication() error {
	return ut.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Delete(applicationName, &metav1.DeleteOptions{})
}

func (ut *UpgradeTest) deleteConfigMap(name, namespace string) error {
	return ut.k8sCli.CoreV1().ConfigMaps(namespace).Delete(name, &metav1.DeleteOptions{})
}

package applicationoperator

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/api/apps/v1"

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
	applicationName            = "test-app-haufmzt"
	eventsServiceDeployment    = "test-app-haufmzt-event-service"
	applicationProxyDeployment = "test-app-haufmzt-application-gateway"
	eventsConfigMapName        = "test-app-haufmzt-configmap-events"
	proxyConfigMapName         = "test-app-haufmzt-configmap-proxy"
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

	if err := ut.createApplication(log); err != nil {
		return errors.Wrap(err, "could not create Application")
	}

	if err := ut.waitForApplication(stop); err != nil {
		return errors.Wrap(err, "could not find resources")
	}

	if err := ut.setVerificationData(integrationNamespace); err != nil {
		return errors.Wrap(err, "could not set verification data")
	}

	log.Info("ApplicationOperator UpgradeTest is set and ready!")
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
			AccessLabel: "app-access-label",
			Description: "Application used by upgradability test",
		},
	})
	return err
}

func (ut *UpgradeTest) waitForApplication(stop <-chan struct{}) error {
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
	application, e := ut.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Get(applicationName, metav1.GetOptions{})

	if e != nil {
		return false, e
	}

	return application.Status.InstallationStatus.Status == "DEPLOYED", nil
}

func (ut *UpgradeTest) setVerificationData(namespace string) error {
	eventsImage, eventsTimestamp, err := ut.getDeploymentData(eventsServiceDeployment, namespace)

	if err != nil {
		return err
	}

	ut.createConfigMap(eventsConfigMapName, namespace, eventsImage, eventsTimestamp)

	proxyImage, proxyTimestamp, e := ut.getDeploymentData(applicationProxyDeployment, namespace)

	if e != nil {
		return e
	}

	ut.createConfigMap(proxyConfigMapName, namespace, proxyImage, proxyTimestamp)

	return nil
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

	_, e := ut.k8sCli.CoreV1().ConfigMaps(namespace).Create(configMap)

	if e != nil {
		return e
	}

	return nil
}

// TestResources tests resources after backup phase
func (ut *UpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Info("ApplicationOperator UpgradeTest testing resources...")

	if err := ut.verifyResources(integrationNamespace); err != nil {
		return errors.Wrap(err, "image versions are not upgraded")
	}

	if err := ut.deleteResources(log, integrationNamespace); err != nil {
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

	configMap, e := ut.getConfigMap(configmapName, namespace)

	if e != nil {
		return e
	}

	previousImage, ok := configMap.Data[imageKey]

	if !ok {
		return errors.New("pre-upgrade image not found")
	}

	previousTimestamp, ok := configMap.Data[timestampKey]

	if !ok {
		return errors.New("pre-upgrade timestamp not found")
	}

	if previousImage != image || previousTimestamp != timestamp.String() {
		return nil
	}

	return errors.New("image and timestamp not changed")
}

func (ut *UpgradeTest) getDeploymentData(name, namespace string) (image string, timestamp metav1.Time, err error) {
	deployment, err := ut.getDeployment(name, namespace)
	if err != nil {
		return "", metav1.Time{}, err
	}

	imageVersion, e := getImageVersion(name, deployment.Spec.Template.Spec.Containers)

	if e != nil {
		return "", metav1.Time{}, e
	}

	return imageVersion, deployment.CreationTimestamp, nil
}

func getImageVersion(containerName string, containers []core.Container) (string, error) {
	for _, c := range containers {
		if strings.Contains(c.Name, containerName) {
			return c.Image, nil
		}
	}
	return "", errors.New(fmt.Sprintf("container name %s not found", containerName))
}

func (ut *UpgradeTest) getDeployment(name, namespace string) (*v1.Deployment, error) {
	return ut.k8sCli.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
}

func (ut *UpgradeTest) getConfigMap(name, namespace string) (*core.ConfigMap, error) {
	return ut.k8sCli.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
}

func (ut *UpgradeTest) deleteResources(log logrus.FieldLogger, namespace string) error {
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

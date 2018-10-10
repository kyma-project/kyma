// +build acceptance

package servicecatalog

import (
	"fmt"
	"log"
	"os"
	"testing"

	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	tester "github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/brokerinstaller"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/k8s"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	brokerReadyTimeout = time.Second * 300
)

func TestMain(m *testing.M) {
	if dex.IsSCIEnabled() {
		log.Println("Skipping configuration, because SCI is enabled")
		return
	}
	k8sClient, _, err := k8s.NewClientWithConfig()
	exitOnError(err, "while initializing K8S Client")

	testClusterBrokerChartPath := fmt.Sprintf("testdata/charts/%s", tester.ClusterBrokerReleaseName)
	testBrokerChartPath := fmt.Sprintf("testdata/charts/%s", tester.BrokerReleaseName)

	serviceCatalogTestPath := os.Getenv("TEST_SERVICE_CATALOG_DIR")
	if serviceCatalogTestPath != "" {
		testClusterBrokerChartPath = fmt.Sprintf("%s/%s", serviceCatalogTestPath, testClusterBrokerChartPath)
		testBrokerChartPath = fmt.Sprintf("%s/%s", serviceCatalogTestPath, testBrokerChartPath)
	}

	clusterBrokerInstaller, err := brokerinstaller.New(testClusterBrokerChartPath, tester.ClusterBrokerReleaseName, tester.DefaultNamespace)
	exitOnError(err, fmt.Sprintf("while initializing installer of %s", tester.ClusterServiceBroker))

	brokerInstaller, err := brokerinstaller.New(testBrokerChartPath, tester.BrokerReleaseName, tester.DefaultNamespace)
	exitOnError(err, fmt.Sprintf("while initializing installer of %s", tester.ServiceBroker))

	err = setup(k8sClient, clusterBrokerInstaller, brokerInstaller)
	if err != nil {
		cleanup(k8sClient, clusterBrokerInstaller, brokerInstaller)
		exitOnError(err, "while setup")
	}

	code := m.Run()

	cleanup(k8sClient, clusterBrokerInstaller, brokerInstaller)
	os.Exit(code)
}

func setup(k8sClient *corev1.CoreV1Client, clusterBrokerInstaller, brokerInstaller *brokerinstaller.BrokerInstaller) error {
	log.Println("Setting up tests...")

	_, err := k8sClient.Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: tester.DefaultNamespace,
		},
	})

	if err != nil {
		return errors.Wrap(err, "while creating namespace")
	}

	svcatCli, _, err := k8s.NewServiceCatalogClientWithConfig()
	if err != nil {
		return errors.Wrap(err, "while initializing service catalog client")
	}

	err = initBroker(clusterBrokerInstaller, svcatCli, tester.ClusterServiceBroker)
	if err != nil {
		return errors.Wrapf(err, "while initializing %s", tester.ClusterServiceBroker)
	}

	err = initBroker(brokerInstaller, svcatCli, tester.ServiceBroker)
	if err != nil {
		return errors.Wrapf(err, "while initializing %s", tester.ServiceBroker)
	}

	return nil
}

func cleanup(k8sClient *corev1.CoreV1Client, clusterBrokerInstaller, brokerInstaller *brokerinstaller.BrokerInstaller) {
	log.Println("Cleaning up...")

	err := clusterBrokerInstaller.Uninstall()
	if err != nil {
		log.Print(errors.Wrapf(err, "while uninstalling chart of %s", tester.ClusterServiceBroker))
	}

	err = brokerInstaller.Uninstall()
	if err != nil {
		log.Print(errors.Wrapf(err, "while uninstalling chart of %s", tester.ServiceBroker))
	}

	err = k8sClient.Namespaces().Delete(tester.DefaultNamespace, nil)
	if err != nil {
		log.Print(errors.Wrap(err, "while deleting namespace"))
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, context))
	}
}

func initBroker(brokerInstaller *brokerinstaller.BrokerInstaller, svcatCli *clientset.Clientset, typeOfBroker string) error {
	err := brokerInstaller.Install()
	if err != nil {
		return errors.Wrapf(err, "while installing %s", typeOfBroker)
	}

	err = waitForBroker(brokerInstaller.ReleaseName(), typeOfBroker, svcatCli)
	if err != nil {
		return errors.Wrapf(err, "while waiting for %s registration", typeOfBroker)
	}

	return nil
}

func checkStatusOfBroker(conditions []v1beta1.ServiceBrokerCondition) bool {
	for _, v := range conditions {
		if v.Type == "Ready" {
			return v.Status == "True"
		}
	}
	return false
}

func waitForBroker(brokerName, typeOfBroker string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		var conditions []v1beta1.ServiceBrokerCondition

		if typeOfBroker == tester.ClusterServiceBroker {
			broker, err := svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Get(brokerName, metav1.GetOptions{})

			if err != nil || broker == nil {
				return false, err
			}

			conditions = broker.Status.Conditions
		} else {
			broker, err := svcatCli.ServicecatalogV1beta1().ServiceBrokers(tester.DefaultNamespace).Get(brokerName, metav1.GetOptions{})

			if err != nil || broker == nil {
				return false, err
			}

			conditions = broker.Status.Conditions
		}

		if checkStatusOfBroker(conditions) {
			return true, nil
		}

		log.Printf("%s %s still not ready. Waiting...\n", typeOfBroker, brokerName)
		return false, nil
	}, brokerReadyTimeout)
}

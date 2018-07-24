// +build acceptance

package servicecatalog

import (
	"fmt"
	"log"
	"os"
	"testing"

	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	tester "github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/brokerinstaller"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/k8s"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const brokerReadyTimeout = time.Second * 300

func TestMain(m *testing.M) {
	k8sClient, _, err := k8s.NewClientWithConfig()
	exitOnError(err, "while initializing K8S Client")

	testBrokerChartPath := "testdata/charts/ups-broker"
	serviceCatalogTestPath := os.Getenv("TEST_SERVICE_CATALOG_DIR")
	if serviceCatalogTestPath != "" {
		testBrokerChartPath = fmt.Sprintf("%s/%s", serviceCatalogTestPath, testBrokerChartPath)
	}

	brokerInstaller, err := brokerinstaller.New(testBrokerChartPath, "ups-broker", tester.DefaultNamespace)
	exitOnError(err, "while initializing installer")

	err = setup(k8sClient, brokerInstaller)
	if err != nil {
		cleanup(k8sClient, brokerInstaller)
		exitOnError(err, "while setup")
	}

	code := m.Run()

	cleanup(k8sClient, brokerInstaller)
	os.Exit(code)
}

func setup(k8sClient *corev1.CoreV1Client, brokerInstaller *brokerinstaller.BrokerInstaller) error {
	log.Println("Setting up tests...")
	_, err := k8sClient.Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: tester.DefaultNamespace,
		},
	})
	if err != nil {
		return errors.Wrap(err, "while creating namespace")
	}

	err = brokerInstaller.Install()
	if err != nil {
		return errors.Wrap(err, "while installing broker")
	}

	svcatCli, _, err := k8s.NewServiceCatalogClientWithConfig()
	if err != nil {
		return errors.Wrap(err, "while initializing service catalog client")
	}

	err = waitForBroker(brokerInstaller.ReleaseName(), svcatCli)
	if err != nil {
		return errors.Wrap(err, "while waiting for broker registration")
	}

	return nil
}

func cleanup(k8sClient *corev1.CoreV1Client, brokerInstaller *brokerinstaller.BrokerInstaller) {
	log.Println("Cleaning up...")
	err := brokerInstaller.Uninstall()
	if err != nil {
		log.Print(errors.Wrap(err, "while uninstalling chart"))
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

func waitForBroker(brokerName string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		broker, err := svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Get(brokerName, metav1.GetOptions{})
		if err != nil || broker == nil {
			return false, err
		}

		arr := broker.Status.Conditions
		for _, v := range arr {
			if v.Type == "Ready" {
				return v.Status == "True", nil
			}
		}

		log.Printf("Broker %s still not ready. Waiting...\n", brokerName)

		return false, nil
	}, brokerReadyTimeout)
}

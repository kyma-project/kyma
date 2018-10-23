// +build acceptance

package servicecatalog

import (
	"log"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/k8s"
	"github.com/pkg/errors"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/installer"
)

const (
	brokerReadyTimeout = time.Second * 300
)

//func TestMain(m *testing.M) {
//	if dex.IsSCIEnabled() {
//		log.Println("Skipping configuration, because SCI is enabled")
//		return
//	}
//	k8sClient, _, err := k8s.NewClientWithConfig()
//	exitOnError(err, "while initializing K8S Client")
//
//	testClusterBrokerChartPath := fmt.Sprintf("testdata/charts/%s", tester.ClusterBrokerReleaseName)
//	testBrokerChartPath := fmt.Sprintf("testdata/charts/%s", tester.BrokerReleaseName)
//
//	serviceCatalogTestPath := os.Getenv("TEST_SERVICE_CATALOG_DIR")
//	if serviceCatalogTestPath != "" {
//		testClusterBrokerChartPath = fmt.Sprintf("%s/%s", serviceCatalogTestPath, testClusterBrokerChartPath)
//		testBrokerChartPath = fmt.Sprintf("%s/%s", serviceCatalogTestPath, testBrokerChartPath)
//	}
//
//	clusterBrokerInstaller, err := brokerinstaller.New(testClusterBrokerChartPath, tester.ClusterBrokerReleaseName, tester.DefaultNamespace)
//	exitOnError(err, fmt.Sprintf("while initializing installer of %s", tester.ClusterServiceBroker))
//
//	brokerInstaller, err := brokerinstaller.New(testBrokerChartPath, tester.BrokerReleaseName, tester.DefaultNamespace)
//	exitOnError(err, fmt.Sprintf("while initializing installer of %s", tester.ServiceBroker))
//
//	err = setup(k8sClient, clusterBrokerInstaller, brokerInstaller)
//	if err != nil {
//		cleanup(k8sClient, clusterBrokerInstaller, brokerInstaller)
//		exitOnError(err, "while setup")
//	}
//
//	code := m.Run()
//
//	cleanup(k8sClient, clusterBrokerInstaller, brokerInstaller)
//	os.Exit(code)
//}

func TestMain(m *testing.M) {
	if dex.IsSCIEnabled() {
		log.Println("Skipping configuration, because SCI is enabled")
		return
	}
	k8sClient, _, err := k8s.NewClientWithConfig()
	exitOnError(err, "while initializing K8S Client")

	svcatCli, _, err := k8s.NewServiceCatalogClientWithConfig()
	exitOnError(err, "while initializing service catalog client")

	clusterBrokerInstaller := installer.NewBroker(tester.ClusterBrokerReleaseName, tester.DefaultNamespace, tester.ClusterServiceBroker)
	brokerInstaller := installer.NewBroker(tester.BrokerReleaseName, tester.DefaultNamespace, tester.ServiceBroker)

	err = setup(k8sClient, svcatCli, clusterBrokerInstaller, brokerInstaller)
	if err != nil {
		cleanup(k8sClient, clusterBrokerInstaller, brokerInstaller)
		exitOnError(err, "while setup")
	}

	code := m.Run()

	cleanup(k8sClient, clusterBrokerInstaller, brokerInstaller)
	os.Exit(code)
}


func setup(k8sClient *corev1.CoreV1Client, svcatCli *clientset.Clientset, clusterBrokerInstaller, brokerInstaller *installer.BrokerInstaller) error {
	log.Println("Setting up tests...")

	err := installer.CreateNamespace(k8sClient, tester.DefaultNamespace)
	if err != nil {
		return errors.Wrap(err, "while creating namespace")
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

func cleanup(k8sClient *corev1.CoreV1Client, svcatCli *clientset.Clientset, clusterBrokerInstaller, brokerInstaller *installer.BrokerInstaller) {
	log.Println("Cleaning up...")

	err := clusterBrokerInstaller.Uninstall(svcatCli)
	if err != nil {
		log.Print(errors.Wrapf(err, "while uninstalling chart of %s", tester.ClusterServiceBroker))
	}

	err = brokerInstaller.Uninstall(svcatCli)
	if err != nil {
		log.Print(errors.Wrapf(err, "while uninstalling chart of %s", tester.ServiceBroker))
	}

	err = installer.DeleteNamespace(k8sClient, tester.DefaultNamespace)
	if err != nil {
		log.Print(errors.Wrap(err, "while deleting namespace"))
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, context))
	}
}

func initBroker(brokerInstaller *installer.BrokerInstaller, svcatCli *clientset.Clientset) error {
	err := brokerInstaller.Install(svcatCli)
	if err != nil {
		return errors.Wrapf(err, "while installing %s", brokerInstaller.TypeOf())
	}

	err = brokerInstaller.WaitForBrokerRunning(svcatCli)
	if err != nil {
		return errors.Wrapf(err, "while waiting for %s registration", brokerInstaller.TypeOf())
	}

	return nil
}

func cleanupBroker(brokerInstaller *installer.BrokerInstaller, svcatCli *clientset.Clientset) error {
	err = brokerInstaller.Uninstall(svcatCli)
	if err != nil {
		log.Print(errors.Wrapf(err, "while uninstalling %s", brokerInstaller.TypeOf()))
	}

	return nil
}
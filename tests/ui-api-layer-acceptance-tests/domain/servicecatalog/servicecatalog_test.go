// +build acceptance

package servicecatalog

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/client"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/installer"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/upsbroker"
	"github.com/pkg/errors"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type serviceCatalogInstallers struct {
	k8sClient              *corev1.CoreV1Client
	svcatCli               *clientset.Clientset
	podInstaller           *installer.PodInstaller
	serviceInstaller       *installer.ServiceInstaller
	clusterBrokerInstaller *installer.ClusterBrokerInstaller
	brokerInstaller        *installer.BrokerInstaller
}

func TestMain(m *testing.M) {
	if dex.IsSCIEnabled() {
		log.Println("Skipping configuration, because SCI is enabled")
		return
	}

	k8sClient, _, err := client.NewClientWithConfig()
	exitOnError(err, "while initializing K8S Client")

	svcatCli, _, err := client.NewServiceCatalogClientWithConfig()
	exitOnError(err, "while initializing service catalog client")

	podInstaller := installer.NewPod(tester.ReleaseName, tester.DefaultNamespace)
	serviceInstaller := installer.NewService(tester.ReleaseName, tester.DefaultNamespace)
	clusterBrokerInstaller := installer.NewClusterBroker(tester.ClusterBrokerReleaseName, tester.DefaultNamespace)
	brokerInstaller := installer.NewBroker(tester.BrokerReleaseName, tester.DefaultNamespace)

	scInstallers := serviceCatalogInstallers{
		k8sClient:              k8sClient,
		svcatCli:               svcatCli,
		podInstaller:           podInstaller,
		serviceInstaller:       serviceInstaller,
		clusterBrokerInstaller: clusterBrokerInstaller,
		brokerInstaller:        brokerInstaller,
	}

	err = setup(scInstallers)
	if err != nil {
		cleanup(scInstallers)
		exitOnError(err, "while setup")
	}

	code := m.Run()

	cleanup(scInstallers)
	os.Exit(code)
}

func setup(scInstallers serviceCatalogInstallers) error {
	log.Println("Setting up tests...")

	err := installer.CreateNamespace(scInstallers.k8sClient, tester.DefaultNamespace)
	if err != nil {
		return errors.Wrap(err, "while creating namespace")
	}

	err = initPod(scInstallers.podInstaller, scInstallers.k8sClient)
	if err != nil {
		return err
	}

	err = initService(scInstallers.serviceInstaller, scInstallers.k8sClient)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s.%s.svc.cluster.local", tester.ReleaseName, tester.DefaultNamespace)

	err = initClusterBroker(scInstallers.clusterBrokerInstaller, scInstallers.svcatCli, url)
	if err != nil {
		return err
	}

	err = initBroker(scInstallers.brokerInstaller, scInstallers.svcatCli, url)
	if err != nil {
		return err
	}

	return nil
}

func cleanup(scInstallers serviceCatalogInstallers) {
	log.Println("Cleaning up...")

	err := scInstallers.podInstaller.Delete(scInstallers.k8sClient)
	if err != nil {
		log.Print(errors.Wrap(err, "while deleting Pod"))
	}

	err = scInstallers.serviceInstaller.Delete(scInstallers.k8sClient)
	if err != nil {
		log.Print(errors.Wrap(err, "while deleting Service"))
	}

	err = scInstallers.clusterBrokerInstaller.Uninstall(scInstallers.svcatCli)
	if err != nil {
		log.Print(errors.Wrapf(err, "while uninstalling %s", tester.ClusterServiceBroker))
	}

	err = scInstallers.brokerInstaller.Uninstall(scInstallers.svcatCli)
	if err != nil {
		log.Print(errors.Wrapf(err, "while uninstalling %s", tester.ServiceBroker))
	}

	err = installer.DeleteNamespace(scInstallers.k8sClient, tester.DefaultNamespace)
	if err != nil {
		log.Print(errors.Wrap(err, "while deleting namespace"))
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, context))
	}
}

func initClusterBroker(clusterBrokerInstaller *installer.ClusterBrokerInstaller, svcatCli *clientset.Clientset, url string) error {
	err := clusterBrokerInstaller.Install(svcatCli, upsbroker.UPSClusterServiceBroker(clusterBrokerInstaller.Name(), url))
	if err != nil {
		return errors.Wrapf(err, "while installing %s", tester.ClusterServiceBroker)
	}

	err = clusterBrokerInstaller.WaitForBrokerRunning(svcatCli)
	if err != nil {
		return errors.Wrapf(err, "while waiting for %s registration", tester.ClusterServiceBroker)
	}

	return nil
}

func initBroker(brokerInstaller *installer.BrokerInstaller, svcatCli *clientset.Clientset, url string) error {
	err := brokerInstaller.Install(svcatCli, upsbroker.UPSServiceBroker(brokerInstaller.Name(), url))
	if err != nil {
		return errors.Wrapf(err, "while installing %s", tester.ServiceBroker)
	}

	err = brokerInstaller.WaitForBrokerRunning(svcatCli)
	if err != nil {
		return errors.Wrapf(err, "while waiting for %s registration", tester.ServiceBroker)
	}

	return nil
}

func initPod(podInstaller *installer.PodInstaller, k8sClient *corev1.CoreV1Client) error {
	err := podInstaller.Create(k8sClient, upsbroker.UPSBrokerPod(podInstaller.Name()))
	if err != nil {
		return errors.Wrapf(err, "while creating Pod")
	}
	err = podInstaller.WaitForPodRunning(k8sClient)
	if err != nil {
		return errors.Wrapf(err, "while waiting for Pod registration")
	}

	return nil
}

func initService(serviceInstaller *installer.ServiceInstaller, k8sClient *corev1.CoreV1Client) error {
	err := serviceInstaller.Create(k8sClient, upsbroker.UPSBrokerService(serviceInstaller.Name()))
	if err != nil {
		return errors.Wrapf(err, "while creating Service")
	}
	err = serviceInstaller.WaitForEndpoint(k8sClient)
	if err != nil {
		return errors.Wrapf(err, "while waiting for Service registration")
	}

	return nil
}

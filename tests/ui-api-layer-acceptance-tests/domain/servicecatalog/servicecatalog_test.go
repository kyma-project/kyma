// +build acceptance

package servicecatalog

import (
	"log"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/k8s"
	"github.com/pkg/errors"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/installer"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"testing"
	"os"
)

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
		cleanup(k8sClient, svcatCli, clusterBrokerInstaller, brokerInstaller)
		exitOnError(err, "while setup")
	}

	code := m.Run()

	cleanup(k8sClient, svcatCli, clusterBrokerInstaller, brokerInstaller)
	os.Exit(code)
}


func setup(k8sClient *corev1.CoreV1Client, svcatCli *clientset.Clientset, clusterBrokerInstaller, brokerInstaller *installer.BrokerInstaller) error {
	log.Println("Setting up tests...")

	err := installer.CreateNamespace(k8sClient, tester.DefaultNamespace)
	if err != nil {
		return errors.Wrap(err, "while creating namespace")
	}

	err = initPod(k8sClient, tester.DefaultNamespace, tester.ReleaseName)
	if err != nil {
		return err
	}

	err = initService(k8sClient, tester.DefaultNamespace, tester.ReleaseName)
	if err != nil {
		return err
	}

	err = initBroker(clusterBrokerInstaller, svcatCli, tester.ReleaseName)
	if err != nil {
		return err
	}

	err = initBroker(brokerInstaller, svcatCli, tester.ReleaseName)
	if err != nil {
		return err
	}

	return nil
}

func cleanup(k8sClient *corev1.CoreV1Client, svcatCli *clientset.Clientset, clusterBrokerInstaller, brokerInstaller *installer.BrokerInstaller) {
	log.Println("Cleaning up...")

	err := installer.DeletePod(k8sClient, tester.DefaultNamespace, tester.ReleaseName)
	if err != nil {
		log.Print(errors.Wrap(err, "while deleting Pod"))
	}

	err = installer.DeleteService(k8sClient, tester.DefaultNamespace, tester.ReleaseName)
	if err != nil {
		log.Print(errors.Wrap(err, "while deleting Service"))
	}

	err = clusterBrokerInstaller.Uninstall(svcatCli)
	if err != nil {
		log.Print(errors.Wrapf(err, "while uninstalling %s", clusterBrokerInstaller.TypeOf()))
	}

	err = brokerInstaller.Uninstall(svcatCli)
	if err != nil {
		log.Print(errors.Wrapf(err, "while uninstalling %s", brokerInstaller.TypeOf()))
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

func initBroker(brokerInstaller *installer.BrokerInstaller, svcatCli *clientset.Clientset, releaseName string) error {
	err := brokerInstaller.Install(svcatCli, releaseName)
	if err != nil {
		return errors.Wrapf(err, "while installing %s", brokerInstaller.TypeOf())
	}

	err = brokerInstaller.WaitForBrokerRunning(svcatCli)
	if err != nil {
		return errors.Wrapf(err, "while waiting for %s registration", brokerInstaller.TypeOf())
	}

	return nil
}

func initPod(k8sClient *corev1.CoreV1Client, namespace, name string) error {
	_, err := installer.CreatePod(k8sClient, namespace, name)
	if err != nil {
		return errors.Wrapf(err, "while creating Pod")
	}
	err = installer.WaitForPodRunning(k8sClient, namespace, name)
	if err != nil {
		return errors.Wrapf(err, "while waiting for Pod registration")
	}

	return nil
}

func initService(k8sClient *corev1.CoreV1Client, namespace, name string) error {
	_, err := installer.CreateService(k8sClient, namespace, name)
	if err != nil {
		return errors.Wrapf(err, "while creating Service")
	}
	err = installer.WaitForEndpoint(k8sClient, namespace, name)
	if err != nil {
		return errors.Wrapf(err, "while waiting for Service registration")
	}

	return nil
}

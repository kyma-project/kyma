package setup

import (
	"fmt"
	"log"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	tester "github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/client"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/installer"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/upsbroker"
	"github.com/pkg/errors"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ServiceCatalogInstaller struct {
	k8sClient              *corev1.CoreV1Client
	svcatCli               *clientset.Clientset
	podInstaller           *installer.PodInstaller
	serviceInstaller       *installer.ServiceInstaller
	clusterBrokerInstaller *installer.ClusterBrokerInstaller
	brokerInstaller        *installer.BrokerInstaller
}

func NewServiceCatalogInstaller() (*ServiceCatalogInstaller, error) {
	k8sClient, _, err := client.NewClientWithConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while initializing K8S Client")
	}

	svcatCli, _, err := client.NewServiceCatalogClientWithConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while initializing service catalog client")
	}

	podInstaller := installer.NewPod(tester.ReleaseName, tester.DefaultNamespace)
	serviceInstaller := installer.NewService(tester.ReleaseName, tester.DefaultNamespace)
	clusterBrokerInstaller := installer.NewClusterBroker(tester.ClusterBrokerReleaseName, tester.DefaultNamespace)
	brokerInstaller := installer.NewBroker(tester.BrokerReleaseName, tester.DefaultNamespace)

	return &ServiceCatalogInstaller{
		k8sClient:              k8sClient,
		svcatCli:               svcatCli,
		podInstaller:           podInstaller,
		serviceInstaller:       serviceInstaller,
		clusterBrokerInstaller: clusterBrokerInstaller,
		brokerInstaller:        brokerInstaller,
	}, nil
}

func (i *ServiceCatalogInstaller) Setup() error {
	log.Println("Setting up tests...")

	err := installer.CreateNamespace(i.k8sClient, tester.DefaultNamespace)
	if err != nil {
		return errors.Wrap(err, "while creating namespace")
	}

	err = i.initPod(i.podInstaller, i.k8sClient)
	if err != nil {
		return err
	}

	err = i.initService(i.serviceInstaller, i.k8sClient)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s.%s.svc.cluster.local", tester.ReleaseName, tester.DefaultNamespace)

	err = i.initClusterBroker(i.clusterBrokerInstaller, i.svcatCli, url)
	if err != nil {
		return err
	}

	err = i.initBroker(i.brokerInstaller, i.svcatCli, url)
	if err != nil {
		return err
	}

	return nil
}

func (i *ServiceCatalogInstaller) Cleanup() {
	log.Println("Cleaning up...")

	err := i.podInstaller.Delete(i.k8sClient)
	if err != nil {
		log.Print(errors.Wrap(err, "while deleting Pod"))
	}

	err = i.serviceInstaller.Delete(i.k8sClient)
	if err != nil {
		log.Print(errors.Wrap(err, "while deleting Service"))
	}

	err = i.clusterBrokerInstaller.Uninstall(i.svcatCli)
	if err != nil {
		log.Print(errors.Wrapf(err, "while uninstalling %s", tester.ClusterServiceBroker))
	}

	err = i.brokerInstaller.Uninstall(i.svcatCli)
	if err != nil {
		log.Print(errors.Wrapf(err, "while uninstalling %s", tester.ServiceBroker))
	}

	err = installer.DeleteNamespace(i.k8sClient, tester.DefaultNamespace)
	if err != nil {
		log.Print(errors.Wrap(err, "while deleting namespace"))
	}
}

func (i *ServiceCatalogInstaller) initClusterBroker(clusterBrokerInstaller *installer.ClusterBrokerInstaller, svcatCli *clientset.Clientset, url string) error {
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

func (i *ServiceCatalogInstaller) initBroker(brokerInstaller *installer.BrokerInstaller,
	svcatCli *clientset.Clientset, url string) error {
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

func (i *ServiceCatalogInstaller) initPod(podInstaller *installer.PodInstaller, k8sClient *corev1.CoreV1Client) error {
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

func (i *ServiceCatalogInstaller) initService(serviceInstaller *installer.ServiceInstaller,
	k8sClient *corev1.CoreV1Client) error {
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

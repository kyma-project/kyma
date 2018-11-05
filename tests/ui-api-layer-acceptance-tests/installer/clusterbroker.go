package installer

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
)

const (
	clusterBrokerReadyTimeout = time.Second * 300
)

type ClusterBrokerInstaller struct {
	name      string
	namespace string
}

func NewClusterBroker(name, namespace string) *ClusterBrokerInstaller {
	return &ClusterBrokerInstaller{
		name:      name,
		namespace: namespace,
	}
}

func (t *ClusterBrokerInstaller) Install(svcatCli *clientset.Clientset, broker *v1beta1.ClusterServiceBroker) error {
	_, err := svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Create(broker)
	return err
}

func (t *ClusterBrokerInstaller) Uninstall(svcatCli *clientset.Clientset) error {
	return svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Delete(t.name, nil)
}

func (t *ClusterBrokerInstaller) Name() string {
	return t.name
}

func (t *ClusterBrokerInstaller) Namespace() string {
	return t.namespace
}

func (t *ClusterBrokerInstaller) WaitForBrokerRunning(svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		broker, err := svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Get(t.name, metav1.GetOptions{})

		if err != nil || broker == nil {
			return false, err
		}
		for _, v := range broker.Status.Conditions {
			if v.Type == "Ready" && v.Status == "True" {
				return true, nil
			}
		}

		return false, fmt.Errorf("%v", broker.Status.Conditions)
	}, clusterBrokerReadyTimeout)
}

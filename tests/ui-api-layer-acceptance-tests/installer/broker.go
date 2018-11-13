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
	brokerReadyTimeout = time.Second * 300
)

type BrokerInstaller struct {
	name      string
	namespace string
}

func NewBroker(name, namespace string) *BrokerInstaller {
	return &BrokerInstaller{
		name:      name,
		namespace: namespace,
	}
}

func (t *BrokerInstaller) Install(svcatCli *clientset.Clientset, broker *v1beta1.ServiceBroker) error {
	_, err := svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.namespace).Create(broker)
	return err
}

func (t *BrokerInstaller) Uninstall(svcatCli *clientset.Clientset) error {
	return svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.namespace).Delete(t.name, nil)
}

func (t *BrokerInstaller) Name() string {
	return t.name
}

func (t *BrokerInstaller) Namespace() string {
	return t.namespace
}

func (t *BrokerInstaller) WaitForBrokerRunning(svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		broker, err := svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.namespace).Get(t.name, metav1.GetOptions{})

		if err != nil || broker == nil {
			return false, err
		}
		for _, v := range broker.Status.Conditions {
			if v.Type == "Ready" && v.Status == "True" {
				return true, nil
			}
		}

		return false, fmt.Errorf("%v", broker.Status.Conditions)
	}, brokerReadyTimeout)
}

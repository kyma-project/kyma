package installer

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	"log"
	"time"
)

const (
	brokerReadyTimeout = time.Second * 300
)

type BrokerInstaller struct {
	releaseName string
	namespace   string
	typeOf 		string
}

func NewBroker(releaseName, namespace, typeOf string) *BrokerInstaller {
	return &BrokerInstaller{
		releaseName: releaseName,
		namespace:   namespace,
		typeOf: 	 typeOf,
	}
}

func (t *BrokerInstaller) Install(svcatCli *clientset.Clientset) error {
	url := "http://" + t.releaseName + "." + t.namespace + ".svc.cluster.local"

	var err error
	if t.typeOf == tester.ClusterServiceBroker {
		_, err = svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Create(upsClusterServiceBroker(t.releaseName, url))
	} else {
		_, err = svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.namespace).Create(upsServiceBroker(t.releaseName, url))
	}
	return err
}

func (t *BrokerInstaller) Uninstall(svcatCli *clientset.Clientset) error {
	var err error
	if t.typeOf == tester.ClusterServiceBroker {
		err = svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Delete(t.releaseName, nil)
	} else {
		err = svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.namespace).Delete(t.releaseName, nil)
	}
	return err
}

func (t *BrokerInstaller) ReleaseName() string {
	return t.releaseName
}

func (t *BrokerInstaller) TypeOf() string {
	return t.typeOf
}

func (t *BrokerInstaller) WaitForBrokerRunning(svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		var conditions []v1beta1.ServiceBrokerCondition

		if t.typeOf == tester.ClusterServiceBroker {
			broker, err := svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Get(t.releaseName, metav1.GetOptions{})

			if err != nil || broker == nil {
				return false, err
			}

			conditions = broker.Status.Conditions
		} else {
			broker, err := svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.namespace).Get(t.releaseName, metav1.GetOptions{})

			if err != nil || broker == nil {
				return false, err
			}

			conditions = broker.Status.Conditions
		}

		if t.checkStatusOfBroker(conditions) {
			return true, nil
		}

		log.Printf("%s %s still not ready. Waiting...\n", t.typeOf, t.releaseName)
		return false, nil
	}, brokerReadyTimeout)
}

func (t *BrokerInstaller) checkStatusOfBroker(conditions []v1beta1.ServiceBrokerCondition) bool {
	for _, v := range conditions {
		if v.Type == "Ready" {
			return v.Status == "True"
		}
	}
	return false
}

func upsClusterServiceBroker(name, url string) *v1beta1.ClusterServiceBroker {
	return &v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}
}

func upsServiceBroker(name, url string) *v1beta1.ServiceBroker {
	return &v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}
}

package nsbroker

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1beta12 "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	brokerName       = "remote-env-broker"
	serviceName      = "service-for-remote-env-broker"
	brokerLabelKey   = "namespaced-remote-env-broker"
	brokerLabelValue = "true"
)

type Manager struct {
	brokerGetter   v1beta12.ServiceBrokersGetter
	servicesGetter v1.ServicesGetter
}

func (m *Manager) Create(ns string) error {
	broker := &v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      brokerName,
			Namespace: ns,
			Labels: map[string]string{
				brokerLabelKey: brokerLabelValue,
			},
		},
	}
	created, err := m.brokerGetter.ServiceBrokers(ns).Create(broker)
	if err != nil {
		return err
	}
	// maybe we should generate some unique prefix
	uuid.NewUUID()
	fmt.Println(created)
	//services
	//service := core.Service{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "",
	//		N
	//	},
	//
	//}

	return nil
}

func (m *Manager) Delete(ns string) error {
	m.brokerGetter.ServiceBrokers(ns).DeleteCollection(nil, )
	err := m.brokerGetter.ServiceBrokers(ns).Delete(brokerName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) Exist(ns string) (bool, error) {
	return false, nil
}

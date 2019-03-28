package configurer

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/waiter"
)

type ServiceBrokerConfig struct {
	Name      string `envconfig:"default=ns-helm-broker"`
	Namespace string`envconfig:"default=default"`
	URL       string `envconfig:"default=http://helm-broker.kyma-system.svc.cluster.local"`
}

type ServiceBrokerConfigurer struct {
	cfg ServiceBrokerConfig

	svcatCli *clientset.Clientset
}

func NewServiceBroker(cfg ServiceBrokerConfig, svcatCli *clientset.Clientset) *ServiceBrokerConfigurer {
	return &ServiceBrokerConfigurer{
		cfg:cfg,
		svcatCli: svcatCli,
	}
}

func (t *ServiceBrokerConfigurer) Create() error {
	broker := &v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.cfg.Name,
			Namespace: t.cfg.Namespace,
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: t.cfg.URL,
			},
		},
	}
	_, err := t.svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.cfg.Namespace).Create(broker)
	return err
}

func (t *ServiceBrokerConfigurer) Delete() error {
	return t.svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.cfg.Namespace).Delete(t.cfg.Name, nil)
}

func (t *ServiceBrokerConfigurer) WaitForReady() error {
	return waiter.WaitAtMost(func() (bool, error) {
		broker, err := t.svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.cfg.Namespace).Get(t.cfg.Name, metav1.GetOptions{})

		if err != nil || broker == nil {
			return false, err
		}
		for _, v := range broker.Status.Conditions {
			if v.Type == "Ready" && v.Status == "True" {
				return true, nil
			}
		}

		return false, nil
	}, brokerReadyTimeout)
}

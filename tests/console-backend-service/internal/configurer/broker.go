package configurer

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/waiter"
)

type ServiceBrokerConfig struct {
	Name      string `envconfig:"default=ns-helm-broker"`
	Namespace string `envconfig:"default=default"`
	URL       string `envconfig:"default=http://helm-broker.kyma-system.svc.cluster.local"`

	ServiceClassExternalName string   `envconfig:"default=testing"`
	ServicePlanExternalNames []string `envconfig:"default=full,minimal"`
}

type ServiceBrokerConfigurer struct {
	cfg ServiceBrokerConfig

	svcatCli *clientset.Clientset
}

func NewServiceBroker(cfg ServiceBrokerConfig, svcatCli *clientset.Clientset) *ServiceBrokerConfigurer {
	return &ServiceBrokerConfigurer{
		cfg:      cfg,
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
	err := t.waitForServiceBrokerReady()
	if err != nil {
		return errors.Wrapf(err, "while waiting for ServiceBroker %s/%s ready", t.cfg.Namespace, t.cfg.Name)
	}

	err = t.waitForServiceClass()
	if err != nil {
		return errors.Wrapf(err, "while waiting for ServiceClass with externalName %s", t.cfg.ServiceClassExternalName)
	}

	err = t.waitForServicePlans()
	if err != nil {
		return errors.Wrapf(err, "while waiting for ServicePlans for ServiceClass with externalName %s", t.cfg.ServiceClassExternalName)
	}

	return nil
}

func (t *ServiceBrokerConfigurer) waitForServiceBrokerReady() error {
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
	}, tester.DefaultReadyTimeout)
}

func (c *ServiceBrokerConfigurer) waitForServiceClass() error {
	return waiter.WaitAtMost(func() (bool, error) {
		classesList, err := c.svcatCli.ServicecatalogV1beta1().ServiceClasses(c.cfg.Namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		for _, class := range classesList.Items {
			if class.GetExternalName() == c.cfg.ServiceClassExternalName {
				return true, nil
			}
		}

		return true, nil
	}, tester.DefaultReadyTimeout)
}

func (c *ServiceBrokerConfigurer) waitForServicePlans() error {
	plansFound := map[string]bool{}

	for _, planName := range c.cfg.ServicePlanExternalNames {
		plansFound[planName] = false
	}

	err := waiter.WaitAtMost(func() (bool, error) {
		planList, err := c.svcatCli.ServicecatalogV1beta1().ServicePlans(c.cfg.Namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		for _, plan := range planList.Items {
			for key := range plansFound {
				if plan.GetExternalName() == key {
					plansFound[key] = true
				}
			}
		}

		for _, value := range plansFound {
			if !value {
				return false, nil
			}
		}

		return true, nil
	}, tester.DefaultReadyTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ServicePlans: %+v", plansFound)
	}

	return nil
}

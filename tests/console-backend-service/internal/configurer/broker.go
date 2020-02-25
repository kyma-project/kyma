package configurer

import (
	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/waiter"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceBrokerConfig struct {
	Name      string `envconfig:"default=ns-helm-broker"`
	Namespace string `envconfig:"default=default"`
	URL       string `envconfig:"default=http://helm-broker.kyma-system.svc.cluster.local/cluster"`

	ServiceClassExternalName string   `envconfig:"default=testing"`
	ServicePlanExternalNames []string `envconfig:"default=full,minimal"`
}

type ServiceBrokerConfigurer struct {
	cfg                 ServiceBrokerConfig
	svcatCli            *clientset.Clientset
	namespaceConfigurer *NamespaceConfigurer
}

func NewServiceBroker(cfg ServiceBrokerConfig, svcatCli *clientset.Clientset, namespaceConfigurer *NamespaceConfigurer) *ServiceBrokerConfigurer {
	return &ServiceBrokerConfigurer{
		cfg:                 cfg,
		svcatCli:            svcatCli,
		namespaceConfigurer: namespaceConfigurer,
	}
}

func (c *ServiceBrokerConfigurer) Create() error {
	broker := &v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.cfg.Name,
			Namespace: c.namespaceConfigurer.Name(),
			Labels: map[string]string{
				tester.TestLabelKey: tester.TestLabelValue,
			},
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: c.cfg.URL,
			},
		},
	}
	_, err := c.svcatCli.ServicecatalogV1beta1().ServiceBrokers(c.namespaceConfigurer.Name()).Create(broker)
	return err
}

func (c *ServiceBrokerConfigurer) Delete() error {
	err := c.svcatCli.ServicecatalogV1beta1().ServiceBrokers(c.namespaceConfigurer.Name()).Delete(c.cfg.Name, nil)
	if err != nil {
		return err
	}

	return c.waitForServiceBrokerDeleted()
}

func (c *ServiceBrokerConfigurer) waitForServiceBrokerDeleted() error {
	err := waiter.WaitAtMost(func() (bool, error) {
		_, err := c.svcatCli.ServicecatalogV1beta1().ServiceBrokers(c.namespaceConfigurer.Name()).Get(c.cfg.Name, metav1.GetOptions{})
		if apiErrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}, tester.DefaultDeletionTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for deletion of ServiceBroker %s", c.cfg.Name)
	}

	return nil
}

func (c *ServiceBrokerConfigurer) WaitForReady() error {
	err := c.waitForServiceBrokerReady()
	if err != nil {
		return errors.Wrapf(err, "while waiting for ServiceBroker %s/%s ready", c.namespaceConfigurer.Name(), c.cfg.Name)
	}

	err = c.waitForServiceClass()
	if err != nil {
		return errors.Wrapf(err, "while waiting for ServiceClass with externalName %s", c.cfg.ServiceClassExternalName)
	}

	err = c.waitForServicePlans()
	if err != nil {
		return errors.Wrapf(err, "while waiting for ServicePlans for ServiceClass with externalName %s", c.cfg.ServiceClassExternalName)
	}

	return nil
}

func (c *ServiceBrokerConfigurer) waitForServiceBrokerReady() error {
	err := waiter.WaitAtMost(func() (bool, error) {
		broker, err := c.svcatCli.ServicecatalogV1beta1().ServiceBrokers(c.namespaceConfigurer.Name()).Get(c.cfg.Name, metav1.GetOptions{})

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
	if err != nil {
		return errors.Wrapf(err, "while waiting for ServiceBroker ready")
	}

	return nil
}

func (c *ServiceBrokerConfigurer) waitForServiceClass() error {
	err := waiter.WaitAtMost(func() (bool, error) {
		classesList, err := c.svcatCli.ServicecatalogV1beta1().ServiceClasses(c.namespaceConfigurer.Name()).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		for _, class := range classesList.Items {
			if class.GetExternalName() == c.cfg.ServiceClassExternalName {
				return true, nil
			}
		}

		return false, nil
	}, tester.DefaultReadyTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ServiceClass")
	}

	return nil
}

func (c *ServiceBrokerConfigurer) waitForServicePlans() error {
	plansFound := map[string]bool{}

	for _, planName := range c.cfg.ServicePlanExternalNames {
		plansFound[planName] = false
	}

	err := waiter.WaitAtMost(func() (bool, error) {
		planList, err := c.svcatCli.ServicecatalogV1beta1().ServicePlans(c.namespaceConfigurer.Name()).List(metav1.ListOptions{})
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
				// one of required plans hasn't been found
				return false, nil
			}
		}

		// all required plans are ready
		return true, nil
	}, tester.DefaultReadyTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ServicePlans: %+v", plansFound)
	}

	return nil
}

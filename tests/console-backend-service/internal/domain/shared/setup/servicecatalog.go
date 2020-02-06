package setup

import (
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/configurer"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/injector"
)

type ServiceCatalogConfigurerConfig struct {
	ServiceBroker    configurer.ServiceBrokerConfig
	TestingAddonsURL string
}

type ServiceCatalogConfigurer struct {
	nsConfigurer     *configurer.NamespaceConfigurer
	brokerConfigurer *configurer.ServiceBrokerConfigurer
	addonsInjector   *injector.Addons
}

func NewServiceCatalogConfigurer(namespace string, registerServiceBroker bool) (*ServiceCatalogConfigurer, error) {
	var cfg ServiceCatalogConfigurerConfig
	err := envconfig.InitWithPrefix(&cfg, "TEST")
	if err != nil {
		return nil, errors.Wrap(err, "while loading config")
	}

	coreCli, _, err := client.NewClientWithConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while initializing K8S Client")
	}

	svcatCli, _, err := client.NewServiceCatalogClientWithConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while initializing service catalog client")
	}

	nsConfigurer := configurer.NewNamespace(namespace, coreCli)

	aInjector, err := injector.NewAddons("testing-addons-cbs", cfg.TestingAddonsURL)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating the addons configuration injector")
	}

	var brokerConfigurer *configurer.ServiceBrokerConfigurer
	if registerServiceBroker {
		cfg.ServiceBroker.Namespace = namespace
		brokerConfigurer = configurer.NewServiceBroker(cfg.ServiceBroker, svcatCli)
	}

	return &ServiceCatalogConfigurer{
		addonsInjector:   aInjector,
		nsConfigurer:     nsConfigurer,
		brokerConfigurer: brokerConfigurer,
	}, nil
}

func (c *ServiceCatalogConfigurer) Setup() error {
	log.Println("Setting up tests...")

	err := c.nsConfigurer.Create()
	if err != nil {
		return errors.Wrap(err, "while creating namespace")
	}

	if err = c.addonsInjector.InjectClusterAddonsConfiguration(); err != nil {
		return errors.Wrapf(err, "while injecting the addons configuration")
	}

	if c.brokerConfigurer != nil {
		err = c.brokerConfigurer.Create()
		if err != nil {
			return errors.Wrap(err, "while creating ServiceBroker")
		}

		err = c.brokerConfigurer.WaitForReady()
		if err != nil {
			return errors.Wrap(err, "while waiting for ServiceBroker ready")
		}
	}

	return nil
}

func (c *ServiceCatalogConfigurer) Cleanup() error {
	log.Println("Cleaning up...")
	var result *multierror.Error

	if c.brokerConfigurer != nil {
		if err := c.brokerConfigurer.Delete(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	if err := c.nsConfigurer.Delete(); err != nil {
		result = multierror.Append(result, err)
	}

	if err := c.addonsInjector.CleanupClusterAddonsConfiguration(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

package setup

import (
	"github.com/vrischmann/envconfig"
	"log"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/configurer"
	"github.com/pkg/errors"
)

type ServiceCatalogConfigurer struct {
	nsConfigurer *configurer.NamespaceConfigurer
	bundleConfigurer *configurer.TestBundleConfigurer
}

func NewServiceCatalogConfigurer(namespace string) (*ServiceCatalogConfigurer, error) {
	var cfg configurer.TestBundleConfig
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

	bundleConfigurer := configurer.NewTestBundle(cfg, coreCli, svcatCli)

	return &ServiceCatalogConfigurer{
		bundleConfigurer: bundleConfigurer,
		nsConfigurer:nsConfigurer,
	}, nil
}

func (i *ServiceCatalogConfigurer) Setup() error {
	log.Println("Setting up tests...")

	err := i.nsConfigurer.Create()
	if err != nil {
		return errors.Wrap(err, "while creating namespace")
	}

	err = i.bundleConfigurer.Configure()
	if err != nil {
		return errors.Wrap(err, "while configuring test bundle")
	}

	err = i.bundleConfigurer.WaitForTestBundleReady()
	if err != nil {
		return errors.Wrap(err, "while waiting for test bundle ready")
	}

	return nil
}

func (i *ServiceCatalogConfigurer) Cleanup() error {
	log.Println("Cleaning up...")

	err := i.bundleConfigurer.Cleanup()
	if err != nil {
		return errors.Wrap(err, "while cleaning up test bundle configuration")
	}

	err = i.nsConfigurer.Delete()
	if err != nil {
		return errors.Wrap(err, "while deleting namespace")
	}

	return nil
}

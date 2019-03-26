package setup

import (
	"fmt"
	"log"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/configurer"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/upsbroker"
	"github.com/pkg/errors"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type TestBundleInstaller struct {
	nsConfigurer *configurer.NamespaceConfigurer
	bundleConfigurer *configurer.TestBundleConfigurer
}

func NewServiceCatalogConfigurer(namespace string) (*TestBundleInstaller, error) {
	k8sClient, _, err := client.NewClientWithConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while initializing K8S Client")
	}

	svcatCli, _, err := client.NewServiceCatalogClientWithConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while initializing service catalog client")
	}

	nsConfigurer := configurer.NewNamespace(namespace, k8sClient)

	bundleConfigurer := configurer.NewTestBundle()

	return &TestBundleInstaller{
		bundleConfigurer: bundleConfigurer,
		nsConfigurer:nsConfigurer,
	}, nil
}

func (i *TestBundleInstaller) Setup() error {
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

func (i *TestBundleInstaller) Cleanup() error {
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

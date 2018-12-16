package application

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/acceptance/application/suite"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	KeepTestResources bool `envconfig:"APPLICATION_KEEP_RESOURCES"`
	Disabled          bool `envconfig:"APPLICATION_DISABLED"`
}

func TestApplicationMapping_EnsureBrokerAndClassesOnlyWhenMappingExist(t *testing.T) {
	var cfg Config
	if err := envconfig.Init(&cfg); err != nil {
		t.Fatalf(err.Error())
	}
	if cfg.Disabled {
		t.Skip("Test skipped due to test configuration.")
	}

	ts := suite.NewMappingTestSuite(t)
	ts.Setup()
	if !cfg.KeepTestResources {
		defer ts.TearDown()
	}

	// Service Broker and Service Class shouldn't exist without ApplicationMapping
	ts.EnsureServiceBrokerNotExist(ts.MappedNs)
	ts.EnsureServiceClassNotExist(ts.MappedNs)

	ts.CreateApplicationMapping()

	t.Log("Waiting for service broker")
	ts.WaitForServiceBrokerWithTimeout(time.Second * 30)

	t.Log("Waiting for service class")
	// timeout must be greater than service broker relist duration
	ts.WaitForServiceClassWithTimeout(time.Second * 90)

	// Service Broker and Service Class shouldn't exist in any other namespace
	ts.EnsureServiceBrokerNotExist(ts.EmptyNs)
	ts.EnsureServiceClassNotExist(ts.EmptyNs)

	ts.DeleteApplicationMapping()

	// Service Broker and Service Class shouldn't exist without ApplicationMapping
	t.Log("Waiting until AppBroker will delete service broker")
	ts.EnsureServiceBrokerNotExistWithTimeout(time.Second * 30)
	ts.EnsureServiceClassNotExist(ts.MappedNs)

}

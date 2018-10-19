package remote_environment

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/acceptance/remote-environment/suite"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	KeepTestResources bool `envconfig:"REMOTE_ENVIRONMENT_KEEP_RESOURCES"`
	Disabled          bool `envconfig:"REMOTE_ENVIRONMENT_DISABLED"`
}

func TestEnvironmentMapping_EnsureBrokerAndClassesOnlyWhenMappingExist(t *testing.T) {
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

	// Service Broker and Service Class shouldn't exist without EnvironmentMapping
	ts.EnsureServiceBrokerNotExist(ts.MappedNs)
	ts.EnsureServiceClassNotExist(ts.MappedNs)

	ts.CreateEnvironmentMapping()

	t.Log("Waiting for service broker")
	ts.WaitForServiceBrokerWithTimeout(time.Second * 30)

	t.Log("Waiting for service class")
	// timeout must be greater than service broker relist duration
	ts.WaitForServiceClassWithTimeout(time.Second * 90)

	// Service Broker and Service Class shouldn't exist in any other namespace
	ts.EnsureServiceBrokerNotExist(ts.EmptyNs)
	ts.EnsureServiceClassNotExist(ts.EmptyNs)

	ts.DeleteEnvironmentMapping()

	// Service Broker and Service Class shouldn't exist without EnvironmentMapping
	t.Log("Waiting until REB will delete service broker")
	ts.EnsureServiceBrokerNotExistWithTimeout(time.Second * 30)
	ts.EnsureServiceClassNotExist(ts.MappedNs)

}

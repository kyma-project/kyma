package remote_environment_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/acceptance/remote-environment/suite"
	"github.com/vrischmann/envconfig"
)

// Config contains all configurations for Remote Environment Acceptance tests
type Config struct {
	DockerImage            string        `envconfig:"STUBS_DOCKER_IMAGE"`
	LinkingTimeout         time.Duration `envconfig:"default=3m,REMOTE_ENVIRONMENT_LINKING_TIMEOUT"`
	UnlinkingTimeout       time.Duration `envconfig:"default=3m,REMOTE_ENVIRONMENT_UNLINKING_TIMEOUT"`
	KeepTestResources      bool          `envconfig:"REMOTE_ENVIRONMENT_KEEP_RESOURCES"`
	Disabled               bool          `envconfig:"REMOTE_ENVIRONMENT_DISABLED"`
	TearDownTimeoutPerStep time.Duration `envconfig:"default=2m,TEAR_DOWN_TIMEOUT_PER_STEP"`
}

func TestRemoteEnvironmentAPIAccess(t *testing.T) {
	var cfg Config
	if err := envconfig.Init(&cfg); err != nil {
		t.Fatalf(err.Error())
	}

	if cfg.Disabled {
		t.Skip("Test skipped due to test configuration.")
	}

	t.Logf("Running Remote Environment Test with config: %+v", cfg)

	// GIVEN
	ts := suite.NewTestSuite(t, cfg.DockerImage, "acceptance-test")
	ts.Setup()
	if !cfg.KeepTestResources {
		defer ts.TearDown(cfg.TearDownTimeoutPerStep)
	}

	t.Logf("Waiting for service class")
	// timeout must be greater than service broker relist duration
	ts.WaitForServiceClassWithTimeout(time.Second * 90)

	t.Logf("Provisioning service instance")
	ts.ProvisionServiceInstance(10 * time.Second)

	t.Logf("Binding")
	ts.Bind(10 * time.Second)

	ts.WaitForPodsAreRunning(45 * time.Second)
	t.Logf("All pods are running")

	t.Logf("Creating binding usage")
	ts.CreateTesterBindingUsage()

	// WHEN / THEN

	t.Logf("Expecting call succeeded and env injected")
	ts.WaitForCallSucceededAndEnvInjected(t, cfg.LinkingTimeout)
	t.Logf("Call succeeded")

	t.Logf("Delete binding usage")
	ts.DeleteTesterBindingUsage()

	t.Logf("Expecting istio forbidden and env not injected")
	ts.WaitForCallForbiddenAndEnvNotInjected(t, cfg.UnlinkingTimeout)
}

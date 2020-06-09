package connectivity_adapter

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

type Scenario struct {
	domain           string
	testID           string
	skipSSLVerify    bool
	eventSendDelay   time.Duration
	testServiceImage string
}

// AddFlags adds CLI flags to given FlagSet
func (s *Scenario) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "connectivity-adapter-e2e", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.DurationVar(&s.eventSendDelay, "eventSendDelay", time.Duration(10)*time.Second, "Wait time in seconds before sending the first event, e.g. 5s")
	pflag.StringVar(&s.testServiceImage, "testServiceImage", "eu.gcr.io/kyma-project/event-subscriber-tools:PR-8483", "TestServiceImage")
}

func (s *Scenario) NewState() (*state, error) {
	config := scenario.CompassEnvConfig{}
	err := envconfig.Init(&config)
	if err != nil {
		return nil, errors.Wrap(err, "while loading environment variables")
	}
	return &state{
		E2EState:         scenario.E2EState{Domain: s.domain, SkipSSLVerify: s.skipSSLVerify, AppName: s.testID, GatewaySubdomain: "adapter-gateway-mtls"},
		CompassEnvConfig: config,
	}, nil
}

func (s *Scenario) RunnerOpts() []step.RunnerOption {
	return nil
}

package compass_e2e

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/vrischmann/envconfig"
)

// CompassE2EScenario executes complete external solution integration test scenario
// using Compass for Application registration and connectivity
type CompassE2EScenario struct {
	domain        string
	testID        string
	skipSSLVerify bool
}

// AddFlags adds CLI flags to given FlagSet
func (s *CompassE2EScenario) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "compass-e2e-test", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
}

func (s *CompassE2EScenario) NewState() (*compassE2EState, error) {
	config := scenario.CompassEnvConfig{}
	err := envconfig.Init(&config)
	if err != nil {
		return nil, errors.Wrap(err, "while loading environment variables")
	}

	return &compassE2EState{
		E2EState:         scenario.E2EState{Domain: s.domain, SkipSSLVerify: s.skipSSLVerify, AppName: s.testID},
		CompassEnvConfig: config,
	}, nil
}

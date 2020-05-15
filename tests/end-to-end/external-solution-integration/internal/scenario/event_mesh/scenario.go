package event_mesh

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// E2EScenario executes complete external solution integration test scenario
type Scenario struct {
	domain            string
	testID            string
	skipSSLVerify     bool
	applicationTenant string
	applicationGroup  string
	waitTime          time.Duration
}

// AddFlags adds CLI flags to given FlagSet
func (s *Scenario) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "e2e-mesh-ns", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.applicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.applicationGroup, "applicationGroup", "", "Application CR Group")
	s.waitTime = time.Duration(*pflag.Int("waitTime", 10, "Wait time in seconds")) * time.Second
}

func (s *Scenario) NewState() *state {
	return &state{
		E2EState: scenario.E2EState{Domain: s.domain, SkipSSLVerify: s.skipSSLVerify, AppName: s.testID, GatewaySubdomain: "gateway"},
	}
}

func (s *Scenario) RunnerOpts() []step.RunnerOption {
	return nil
}

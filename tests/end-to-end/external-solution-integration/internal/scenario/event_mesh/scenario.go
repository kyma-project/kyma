package event_mesh

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
)

// E2EScenario executes complete external solution integration test scenario
type Scenario struct {
	domain            string
	testID            string
	skipSSLVerify     bool
	applicationTenant string
	applicationGroup  string
	waitTime          time.Duration
	logLevel          string
}

// AddFlags adds CLI flags to given FlagSet
func (s *Scenario) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "e2e-mesh-ns", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.applicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.applicationGroup, "applicationGroup", "", "Application CR Group")
	pflag.DurationVar(&s.waitTime, "waitTime", time.Duration(10)*time.Second, "Wait time in seconds, e.g. 5s")
	pflag.StringVar(&s.logLevel, "logLevel", "info", "Set log level: panic, fatal, error, warn, info, debug, trace")
}

func (s *Scenario) NewState() *state {
	return &state{
		E2EState: scenario.E2EState{Domain: s.domain, SkipSSLVerify: s.skipSSLVerify, AppName: s.testID, GatewaySubdomain: "gateway"},
	}
}

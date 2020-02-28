package e2e

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/spf13/pflag"
)

// E2EScenario executes complete external solution integration test scenario
type E2EScenario struct {
	Domain            string
	testID            string
	SkipSSLVerify     bool
	ApplicationTenant string
	ApplicationGroup  string
}

// AddFlags adds CLI flags to given FlagSet
func (s *E2EScenario) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.Domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "e2e-test", "domain")
	pflag.BoolVar(&s.SkipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.ApplicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.ApplicationGroup, "applicationGroup", "", "Application CR Group")
}

func (s *E2EScenario) NewState() *e2EState {
	return &e2EState{E2EState: scenario.E2EState{Domain: s.Domain, SkipSSLVerify: s.SkipSSLVerify, AppName: s.testID}}
}

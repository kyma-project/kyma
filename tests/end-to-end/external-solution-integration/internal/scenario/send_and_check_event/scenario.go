package send_and_check_event

import (
	"github.com/spf13/pflag"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// Scenario executes complete external solution integration test scenario
type Scenario struct {
	Domain            string
	testID            string
	SkipSSLVerify     bool
	ApplicationTenant string
	ApplicationGroup  string
	testServiceImage  string
}

// AddFlags adds CLI flags to given FlagSet
func (s *Scenario) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.Domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "e2e-test", "domain")
	pflag.BoolVar(&s.SkipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.ApplicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.ApplicationGroup, "applicationGroup", "", "Application CR Group")
	pflag.StringVar(&s.testServiceImage, "testServiceImage", "eu.gcr.io/kyma-project/event-subscriber-tools:PR-8483", "TestServiceImage")
}

func (s *Scenario) NewState() *state {
	return &state{E2EState: scenario.E2EState{Domain: s.Domain, SkipSSLVerify: s.SkipSSLVerify, AppName: s.testID, GatewaySubdomain: "gateway"}}
}

func (s *Scenario) RunnerOpts() []step.RunnerOption {
	return nil
}

package event_mesh_prepare

import (
	"github.com/spf13/pflag"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// E2EScenario executes complete external solution integration test scenario
type Scenario struct {
	Domain            string
	TestID            string
	SkipSSLVerify     bool
	ApplicationTenant string
	ApplicationGroup  string
}

// AddFlags adds CLI flags to given FlagSet
func (s *Scenario) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.Domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.TestID, "testID", "external-solution-test", "domain")
	pflag.BoolVar(&s.SkipSSLVerify, "SkipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.ApplicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.ApplicationGroup, "applicationGroup", "", "Application CR Group")
}

func (s *Scenario) NewState() *e2EEventMeshState {
	return &e2EEventMeshState{
		E2EState: scenario.E2EState{Domain: s.Domain, SkipSSLVerify: s.SkipSSLVerify, AppName: s.TestID, GatewaySubdomain: "gateway"},
	}
}

func (s *Scenario) RunnerOpts() []step.RunnerOption {
	return []step.RunnerOption{
		step.WithCleanupDefault(step.CleanupModeOnErrorOnly),
	}
}

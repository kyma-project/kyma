package event_mesh_evaluate

import (
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
}

// AddFlags adds CLI flags to given FlagSet
func (s *Scenario) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "external-solution-test", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.applicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.applicationGroup, "applicationGroup", "", "Application CR Group")
}

func (s *Scenario) NewState() *e2EEventMeshState {
	return &e2EEventMeshState{
		E2EState: scenario.E2EState{Domain: s.domain, SkipSSLVerify: s.skipSSLVerify, AppName: s.testID, GatewaySubdomain: "gateway"},
	}
}

func (s *Scenario) RunnerOpts() []step.RunnerOption {
	return nil
}

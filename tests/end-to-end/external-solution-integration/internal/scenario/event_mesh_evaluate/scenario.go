package event_mesh_evaluate

import (
	"github.com/spf13/pflag"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// Scenario executes the evaluation of the 2 phase end to end scenario
type Scenario struct {
	Domain        string
	TestID        string
	SkipSSLVerify bool
}

// AddFlags adds CLI flags to given FlagSet
func (s *Scenario) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.Domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.TestID, "testID", "external-solution-test", "testID")
	pflag.BoolVar(&s.SkipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
}

func (s *Scenario) NewState() *e2EEventMeshState {
	return &e2EEventMeshState{
		E2EState: scenario.E2EState{Domain: s.Domain, SkipSSLVerify: s.SkipSSLVerify, AppName: s.TestID, GatewaySubdomain: "gateway"},
	}
}

func (s *Scenario) RunnerOpts() []step.RunnerOption {
	return nil
}

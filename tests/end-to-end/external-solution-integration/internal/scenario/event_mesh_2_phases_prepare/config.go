package event_mesh_2_phases_prepare

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/spf13/pflag"
)

// E2EScenario executes complete external solution integration test scenario
type TwoPhasesEventMeshPrepareConfig struct {
	domain            string
	testID            string
	skipSSLVerify     bool
	applicationTenant string
	applicationGroup  string
}

// AddFlags adds CLI flags to given FlagSet
func (s *TwoPhasesEventMeshPrepareConfig) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "external-solution-test", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.applicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.applicationGroup, "applicationGroup", "", "Application CR Group")
}

func (s *TwoPhasesEventMeshPrepareConfig) NewState() *e2EEventMeshState {
	return &e2EEventMeshState{
		E2EState: scenario.E2EState{Domain: s.domain, SkipSSLVerify: s.skipSSLVerify, AppName: s.testID},
	}
}

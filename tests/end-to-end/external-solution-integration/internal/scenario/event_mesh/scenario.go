package event_mesh

import (
	"github.com/spf13/pflag"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_evaluate"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_prepare"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// E2EScenario executes complete external solution integration test scenario
type Scenario struct {
	prepare           event_mesh_prepare.Scenario
	evaluate          event_mesh_evaluate.Scenario
	domain            string
	testID            string
	skipSSLVerify     bool
	applicationTenant string
	applicationGroup  string
	logLevel          string
}

// AddFlags adds CLI flags to given FlagSet
func (s *Scenario) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "external-solution-test", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.applicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.applicationGroup, "applicationGroup", "", "Application CR Group")
	pflag.StringVar(&s.logLevel, "logLevel", "info", "Set log level: panic, fatal, error, warn, info, debug, trace")
}

func (s *Scenario) NewState() *state {
	return &state{}
}

func (s *Scenario) RunnerOpts() []step.RunnerOption {
	runnerOpts := s.prepare.RunnerOpts()
	runnerOpts = append(runnerOpts, s.evaluate.RunnerOpts()...)
	return append(runnerOpts, step.WithCleanupDefault(step.CleanupModeYes))
}

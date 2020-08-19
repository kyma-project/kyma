package scenario

import (
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// Scenario represents a test scenario to be run
type Scenario interface {
	AddFlags(set *pflag.FlagSet)
	Steps(config *rest.Config) ([]step.Step, error)
	RunnerOpts() []step.RunnerOption
}

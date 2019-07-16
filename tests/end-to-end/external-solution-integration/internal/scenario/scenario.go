package scenario

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
)

type Scenario interface {
	AddFlags(set *pflag.FlagSet)
	Steps(config *rest.Config) ([]step.Step, error)
}

package teststep

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
)

type APIRule struct {
	apirule     *apirule.APIRule
	stepName    string
	serviceName string
	domainName  string
	domainPort  uint32
}

func (A APIRule) Name() string {
	return A.stepName
}

func (A APIRule) Run() error {
	domainHost := fmt.Sprintf("%s.%s", A.serviceName, A.domainName)
	if _, err := A.apirule.Create(A.serviceName, domainHost, A.domainPort); err != nil {
		return err
	}

	return A.apirule.WaitForStatusRunning()
}

func (A APIRule) Cleanup() error {
	return A.apirule.Delete()
}

func NewAPIRule(rule *apirule.APIRule, stepName, serviceName, domainName string, domainPort uint32) APIRule {
	return APIRule{
		apirule:     rule,
		stepName:    stepName,
		serviceName: serviceName,
		domainName:  domainName,
		domainPort:  domainPort,
	}
}

var _ step.Step = APIRule{}

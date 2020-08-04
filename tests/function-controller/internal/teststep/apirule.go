package teststep

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
)

type APIRule struct {
	apirule     *apirule.APIRule
	name        string
	serviceName string
	domainName  string
	domainPort  uint32
}

func (A APIRule) Name() string {
	return A.name
}

func (A APIRule) Run() error {
	//            "host": "lucky-cancer.wookiee-2596996162.hudy.ninja",
	//    host: test.lucky-cancer.wookiee.hudy.ninja
	//domainHost := fmt.Sprintf("%s-%d.%s", t.cfg.DomainName, rand.Uint32(), t.cfg.domainName)

	domainHost := fmt.Sprintf("%s.%s", A.serviceName, A.domainName)
	if _, err := A.apirule.Create(A.serviceName, domainHost, A.domainPort); err != nil {
		return err
	}

	return A.apirule.WaitForStatusRunning()
}

func (A APIRule) Cleanup() error {
	return nil
}

func NewAPIRule(rule *apirule.APIRule, name, serviceName, domainName string, domainPort uint32) APIRule {
	return APIRule{
		apirule:     rule,
		name:        name,
		serviceName: serviceName,
		domainName:  domainName,
		domainPort:  domainPort,
	}
}

var _ step.Step = APIRule{}

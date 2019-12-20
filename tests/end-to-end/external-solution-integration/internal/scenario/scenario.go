package scenario

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
)

// Scenario represents a test scenario to be run
type Scenario interface {
	AddFlags(set *pflag.FlagSet)
	Steps(config *rest.Config) ([]step.Step, error)
}

type e2eState struct {
	domain        string
	skipSSLVerify bool
	appName       string

	apiServiceInstanceName   string
	eventServiceInstanceName string
	eventSender              *testkit.EventSender
}

// SetAPIServiceInstanceName allows to set APIServiceInstanceName so it can be shared between steps
func (s *e2eState) SetAPIServiceInstanceName(serviceID string) {
	s.apiServiceInstanceName = serviceID
}

// SetEventServiceInstanceName allows to set EventServiceInstanceName so it can be shared between steps
func (s *e2eState) SetEventServiceInstanceName(serviceID string) {
	s.eventServiceInstanceName = serviceID
}

// GetAPIServiceInstanceName allows to get APIServiceInstanceName so it can be shared between steps
func (s *e2eState) GetAPIServiceInstanceName() string {
	return s.apiServiceInstanceName
}

// GetEventServiceInstanceName allows to get EventServiceInstanceName so it can be shared between steps
func (s *e2eState) GetEventServiceInstanceName() string {
	return s.eventServiceInstanceName
}

// GetEventSender returns connected EventSender
func (s *e2eState) GetEventSender() *testkit.EventSender {
	return s.eventSender
}

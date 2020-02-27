package scenario

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type E2EState struct {
	Domain        string
	SkipSSLVerify bool
	AppName       string

	ApiServiceInstanceName   string
	EventServiceInstanceName string
	EventSender              *testkit.EventSender
}

// SetAPIServiceInstanceName allows to set APIServiceInstanceName so it can be shared between steps
func (s *E2EState) SetAPIServiceInstanceName(serviceID string) {
	s.ApiServiceInstanceName = serviceID
}

// SetEventServiceInstanceName allows to set EventServiceInstanceName so it can be shared between steps
func (s *E2EState) SetEventServiceInstanceName(serviceID string) {
	s.EventServiceInstanceName = serviceID
}

// GetAPIServiceInstanceName allows to get APIServiceInstanceName so it can be shared between steps
func (s *E2EState) GetAPIServiceInstanceName() string {
	return s.ApiServiceInstanceName
}

// GetEventServiceInstanceName allows to get EventServiceInstanceName so it can be shared between steps
func (s *E2EState) GetEventServiceInstanceName() string {
	return s.EventServiceInstanceName
}

// GetEventSender returns connected EventSender
func (s *E2EState) GetEventSender() *testkit.EventSender {
	return s.EventSender
}

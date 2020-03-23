package scenario

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type E2EState struct {
	Domain        string
	SkipSSLVerify bool
	AppName       string

	ServiceClassID           string
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

// SetServiceClassID allows to set ServiceClassID so it can be shared between steps
func (s *E2EState) SetServiceClassID(serviceID string) {
	s.ServiceClassID = serviceID
}

// GetServiceClassID allows to get ServiceClassID so it can be shared between steps
func (s *E2EState) GetServiceClassID() string {
	return s.ServiceClassID
}

// GetEventSender returns connected EventSender
func (s *E2EState) GetEventSender() *testkit.EventSender {
	return s.EventSender
}

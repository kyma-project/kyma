package connectivity_adapter

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type state struct {
	scenario.E2EState
	scenario.CompassEnvConfig
	compassAppID  string
	servicePlanID string
}

func (s *state) GetEventSender() *testkit.EventSender {
	return s.EventSender
}

// SetCompassAppID sets Compass ID of registered application
func (s *state) SetCompassAppID(appID string) {
	s.compassAppID = appID
}

// GetCompassAppID returns Compass ID of registered application
func (s *state) GetCompassAppID() string {
	return s.compassAppID
}

func (s *state) GetServicePlanID() string {
	return s.servicePlanID
}

func (s *state) SetServicePlanID(servicePlanID string) {
	s.servicePlanID = servicePlanID
}

package compass

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
)

type state struct {
	scenario.E2EState
	scenario.CompassEnvConfig
	compassAppID  string
	servicePlanID string
}

// GetCompassAppID returns Compass ID of registered application
func (s *state) GetCompassAppID() string {
	return s.compassAppID
}

// SetCompassAppID sets Compass ID of registered application
func (s *state) SetCompassAppID(appID string) {
	s.compassAppID = appID
}

func (s *state) GetServicePlanID() string {
	return s.servicePlanID
}

func (s *state) SetServicePlanID(servicePlanID string) {
	s.servicePlanID = servicePlanID
}

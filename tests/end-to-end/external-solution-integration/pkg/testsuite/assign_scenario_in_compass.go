package testsuite

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

const (
	defaultScenarioName = "DEFAULT"
)

// AssignScenarioInCompass is a step which assigns default scenario to Runtime in Compass
type AssignScenarioInCompass struct {
	name      string
	runtimeID string
	director  *testkit.CompassDirectorClient
}

var _ step.Step = &AssignScenarioInCompass{}

// NewAssignScenarioInCompass returns new AssignScenarioInCompass
func NewAssignScenarioInCompass(name, runtimeID string, director *testkit.CompassDirectorClient) *AssignScenarioInCompass {
	return &AssignScenarioInCompass{
		name:      name,
		runtimeID: runtimeID,
		director:  director,
	}
}

// Name returns name of the step
func (s *AssignScenarioInCompass) Name() string {
	return "Assign default scenario label to runtime in Compass"
}

// Run executes the step
func (s *AssignScenarioInCompass) Run() error {
	return s.director.AddScenarioToRuntime(s.runtimeID, defaultScenarioName)
}

// Cleanup removes all resources that may possibly created by the step
func (s *AssignScenarioInCompass) Cleanup() error {
	return s.director.RemoveScenarioFromRuntime(s.runtimeID, defaultScenarioName)
}

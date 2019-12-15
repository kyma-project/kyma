package testsuite

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type AssignScenarioInCompass struct {
	name        string
	runtimeID string
	scenarioName string
	director *testkit.CompassDirectorClient
	defaultScenarioName string
}

var _ step.Step = &AssignScenarioInCompass{}

func NewAssignScenarioInCompass(name, runtimeID string,  director *testkit.CompassDirectorClient) *AssignScenarioInCompass {
	return &AssignScenarioInCompass{
		name:        name,
		runtimeID: runtimeID,
		scenarioName: "e2e-test",
		director: director,
		defaultScenarioName: "DEFAULT",
	}
}

func (s *AssignScenarioInCompass) Name() string {
	return "Assign default scenario label to runtime in Compass"
}

func (s *AssignScenarioInCompass) Run() error {
	return s.director.AddScenarioToRuntime(s.runtimeID, s.defaultScenarioName)
}

func (s *AssignScenarioInCompass) Cleanup() error {
	return s.director.RemoveScenarioFromRuntime(s.runtimeID, s.defaultScenarioName)
}

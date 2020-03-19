package testsuite

import (
	"encoding/json"
	"fmt"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/example_schema"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/pkg/errors"
)

// RegisterApplicationInCompass is a step which registers new Application with API and Event in Compass
type RegisterLegacyServiceInCompass struct {
	name        string
	apiURL      string
	state       RegisterServiceInCompassState
	testService *testkit.TestService
}

type RegisterServiceInCompassState interface {
	GetRegistryClient() *testkit.RegistryClient
	GetCompassAppID() string
	SetServicePlanID(servicePlanID string)
}

var _ step.Step = &RegisterLegacyServiceInCompass{}

// NewRegisterApplicationInCompass returns new RegisterLegacyServiceInCompass
func NewRegisterLegacyServiceInCompass(name, apiURL string, testService *testkit.TestService, state RegisterServiceInCompassState) *RegisterLegacyServiceInCompass {
	return &RegisterLegacyServiceInCompass{
		name:        name,
		apiURL:      apiURL,
		state:       state,
		testService: testService,
	}
}

// Name returns name of the step
func (s *RegisterLegacyServiceInCompass) Name() string {
	return fmt.Sprintf("Register service using legacy connector %s in compass", s.name)
}

// Run executes the step
func (s *RegisterLegacyServiceInCompass) Run() error {
	service := s.prepareService(s.testService.GetInClusterTestServiceURL())
	serviceID, err := s.state.GetRegistryClient().RegisterService(service)
	if err != nil {
		return errors.Wrap(err, "while registering legacy service")
	}

	s.state.SetServicePlanID(serviceID)
	return nil
}

func (s *RegisterLegacyServiceInCompass) prepareService(targetURL string) *testkit.ServiceDetails {
	return &testkit.ServiceDetails{
		Provider:         s.name,
		Name:             s.name,
		Description:      s.name,
		ShortDescription: s.name,
		Identifier:       s.name,
		Events: &testkit.Events{
			Spec: json.RawMessage(example_schema.EventsSpec),
		},
		Api: &testkit.API{
			TargetUrl: targetURL,
		},
	}
}

// Cleanup removes all resources that may possibly created by the step
func (s *RegisterLegacyServiceInCompass) Cleanup() error {
	return nil
}

type LegacyServiceLabel struct {
	ID         string `json:"id"`
	APIDefID   string `json:"apiDefID"`
	EventDefID string `json:"eventDefID"`
}

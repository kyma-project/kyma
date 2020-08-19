package testsuite

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

// RegisterLegacyApplicationInCompass is a step which registers new Application with API and Event in Compass
type RegisterLegacyApplicationInCompass struct {
	name     string
	director *testkit.CompassDirectorClient

	state RegisterLegacyApplicationInCompassState
}

// RegisterApplicationInCompassState represents RegisterApplicationInCompass dependencies
type RegisterLegacyApplicationInCompassState interface {
	GetCompassAppID() string
	SetCompassAppID(appID string)
	SetServiceClassID(serviceID string)
}

var _ step.Step = &RegisterLegacyApplicationInCompass{}

// RegisterEmptyApplicationInCompass returns new RegisterLegacyApplicationInCompass
func RegisterEmptyApplicationInCompass(name string, director *testkit.CompassDirectorClient, state RegisterLegacyApplicationInCompassState) *RegisterLegacyApplicationInCompass {
	return &RegisterLegacyApplicationInCompass{
		name:     name,
		director: director,
		state:    state,
	}
}

// Name returns name of the step
func (s *RegisterLegacyApplicationInCompass) Name() string {
	return fmt.Sprintf("Register application %s in compass", s.name)
}

// Run executes the step
func (s *RegisterLegacyApplicationInCompass) Run() error {
	providerName := "external solution company"
	desc := "Test Application"
	appInput := graphql.ApplicationRegisterInput{
		Name:         s.name,
		Description:  &desc,
		ProviderName: &providerName,
	}

	app, err := s.director.RegisterApplication(appInput)
	if err != nil {
		return errors.Wrap(err, "while registration of application in compass")
	}
	if app.ID == "" {
		return errors.New("registered application id not found")
	}
	s.state.SetCompassAppID(app.ID)
	s.state.SetServiceClassID(app.ID)

	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *RegisterLegacyApplicationInCompass) Cleanup() error {
	if s.state.GetCompassAppID() != "" {
		err := retry.Do(func() error {
			_, err := s.director.UnregisterApplication(s.state.GetCompassAppID())
			return err
		})
		return err
	}

	return nil
}

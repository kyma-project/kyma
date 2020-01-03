package testsuite

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	acClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/example_schema"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

// RegisterApplicationInCompass is a step which registers new Application with API and Event in Compass
type RegisterApplicationInCompass struct {
	name         string
	applications acClient.ApplicationInterface
	director     *testkit.CompassDirectorClient

	apiURL string
	state  RegisterApplicationInCompassState
}

// RegisterApplicationInCompassState represents RegisterApplicationInCompass dependencies
type RegisterApplicationInCompassState interface {
	GetCompassAppID() string
	SetCompassAppID(appID string)
}

var _ step.Step = &RegisterApplicationInCompass{}

// NewRegisterApplicationInCompass returns new RegisterApplicationInCompass
func NewRegisterApplicationInCompass(name, apiURL string, applications acClient.ApplicationInterface, director *testkit.CompassDirectorClient, state RegisterApplicationInCompassState) *RegisterApplicationInCompass {
	return &RegisterApplicationInCompass{
		name:         name,
		applications: applications,
		director:     director,
		apiURL:       apiURL,
		state:        state,
	}
}

// Name returns name of the step
func (s *RegisterApplicationInCompass) Name() string {
	return fmt.Sprintf("Register application %s in compass", s.name)
}

// Run executes the step
func (s *RegisterApplicationInCompass) Run() error {
	eventSpec := s.prepareEventSpec(example_schema.EventsSpec)
	appInput := graphql.ApplicationRegisterInput{
		Name:         s.name,
		ProviderName: "external solution company",
		APIDefinitions: []*graphql.APIDefinitionInput{
			{
				Name:      s.name,
				TargetURL: s.apiURL,
			},
		},
		EventDefinitions: []*graphql.EventDefinitionInput{
			{
				Name: s.name,
				Spec: &graphql.EventSpecInput{
					Data:   &eventSpec,
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: graphql.SpecFormatJSON,
				},
			},
		},
	}

	app, err := s.director.RegisterApplication(appInput)
	if err != nil {
		return err
	}
	if app.ID == "" {
		return errors.New("registered application id not found")
	}
	s.state.SetCompassAppID(app.ID)

	return retry.Do(s.isApplicationReady, retry.Delay(time.Second))
}

func (s *RegisterApplicationInCompass) isApplicationReady() error {
	application, err := s.applications.Get(s.name, v1.GetOptions{})
	if err != nil {
		return err
	}

	if application.Status.InstallationStatus.Status == "DEPLOYED" {
		return errors.Errorf("unexpected installation status: %s", application.Status.InstallationStatus.Status)
	}

	return nil
}

func (s *RegisterApplicationInCompass) prepareEventSpec(inSpec string) graphql.CLOB {
	escapedQuotes := strings.Replace(inSpec, `"`, `\"`, -1)
	removedNewLines := strings.Replace(escapedQuotes, "\n", "", -1)
	return graphql.CLOB(removedNewLines)
}

// Cleanup removes all resources that may possibly created by the step
func (s *RegisterApplicationInCompass) Cleanup() error {
	if s.state.GetCompassAppID() != "" {
		_, err := s.director.UnregisterApplication(s.state.GetCompassAppID())
		if err != nil {
			return err
		}
	}

	return nil
}

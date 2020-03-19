package testsuite

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	acclient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	esclient "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/example_schema"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/pkg/errors"
)

// RegisterApplicationInCompass is a step which registers new Application with API and Event in Compass
type RegisterApplicationInCompass struct {
	name         string
	applications acclient.ApplicationInterface
	httpSources  esclient.HTTPSourceInterface
	director     *testkit.CompassDirectorClient

	apiURL string
	state  RegisterApplicationInCompassState
}

// RegisterApplicationInCompassState represents RegisterApplicationInCompass dependencies
type RegisterApplicationInCompassState interface {
	GetCompassAppID() string
	SetCompassAppID(appID string)
	SetServiceClassID(serviceID string)
	SetServicePlanID(servicePlanID string)
}

var _ step.Step = &RegisterApplicationInCompass{}

// NewRegisterApplicationInCompass returns new RegisterApplicationInCompass
func NewRegisterApplicationInCompass(name, apiURL string, applications acclient.ApplicationInterface, director *testkit.CompassDirectorClient, httpSources esclient.HTTPSourceInterface, state RegisterApplicationInCompassState) *RegisterApplicationInCompass {
	return &RegisterApplicationInCompass{
		name:         name,
		applications: applications,
		httpSources:  httpSources,
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
	providerName := "external solution company"
	desc := "Sample description"
	appInput := graphql.ApplicationRegisterInput{
		Name:         s.name,
		Description:  &desc,
		ProviderName: &providerName,
		Packages: []*graphql.PackageCreateInput{
			{
				Name:                s.name,
				Description:         &desc,
				DefaultInstanceAuth: &graphql.AuthInput{},
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
				Documents: nil,
			},
		},
	}

	app, err := s.director.RegisterApplication(appInput)
	if err != nil {
		return err
	}
	if app.ID == "" {
		return errors.New("registered Application ID not found")
	}
	s.state.SetCompassAppID(app.ID)
	s.state.SetServiceClassID(app.ID)

	if app.Packages.Data == nil || len(app.Packages.Data) != 1 {
		return fmt.Errorf("registered Application has %d Packages; should contain exactly one Package", len(app.Packages.Data))
	}
	s.state.SetServicePlanID(app.Packages.Data[0].ID)

	return retry.Do(s.isApplicationReady)
}

func (s *RegisterApplicationInCompass) isApplicationReady() error {
	application, err := s.applications.Get(s.name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if application.Status.InstallationStatus.Status == "DEPLOYED" {
		return errors.Errorf("unexpected installation status: %s", application.Status.InstallationStatus.Status)
	}

	return retry.Do(s.isHttpSourceReady)
}

func (s *RegisterApplicationInCompass) isHttpSourceReady() error {
	httpsource, err := s.httpSources.Get(s.name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if !httpsource.Status.IsReady() {
		return errors.Errorf("httpsource %s is not ready. Status of HttpSource: \n %+v", httpsource.Name, httpsource.Status)
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

package testsuite

import (
	"encoding/json"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/example_schema"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type RegisterTestService struct {
	testService *testkit.TestService
	state       RegisterTestServiceState
	name        string
}

type RegisterTestServiceState interface {
	SetServiceClassID(serviceID string)
	GetServiceClassID() string
	GetRegistryClient() *testkit.RegistryClient
}

var _ step.Step = &RegisterTestService{}

func NewRegisterTestService(name string, testService *testkit.TestService, state RegisterTestServiceState) *RegisterTestService {
	return &RegisterTestService{
		testService: testService,
		state:       state,
		name:        name,
	}
}

func (s *RegisterTestService) Name() string {
	return "Register test service"
}

func (s *RegisterTestService) Run() error {
	url := s.testService.GetInClusterTestServiceURL()
	service := s.prepareService(url)

	id, err := s.state.GetRegistryClient().RegisterService(service)
	if err != nil {
		return err
	}

	s.state.SetServiceClassID(id)
	return nil
}

func (s *RegisterTestService) prepareService(targetURL string) *testkit.ServiceDetails {
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

func (s *RegisterTestService) Cleanup() error {
	if serviceID := s.state.GetServiceClassID(); serviceID != "" {
		return s.state.GetRegistryClient().DeleteService(serviceID)
	}
	return nil
}

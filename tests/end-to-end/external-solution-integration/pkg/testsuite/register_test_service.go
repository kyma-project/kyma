package testsuite

import (
	"encoding/json"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

const (
serviceProvider         = "e2e"
serviceName             = "e2e-test-app-svc"
serviceDescription      = "e2e testing app"
serviceShortDescription = "e2e testing app"
serviceIdentifier       = "e2e-test-app-svc-id"
serviceEventsSpec       = `{
   "asyncapi":"1.0.0",
   "info":{
      "title":"Example Events",
      "version":"1.0.0",
      "description":"Description of all the example events"
   },
   "baseTopic":"example.events.com",
   "topics":{
      "exampleEvent.v1":{
         "subscribe":{
            "summary":"Example event",
            "payload":{
               "type":"object",
               "properties":{
                  "myObject":{
                     "type":"object",
                     "required":[
                        "id"
                     ],
                     "example":{
                        "id":"4caad296-e0c5-491e-98ac-0ed118f9474e"
                     },
                     "properties":{
                        "id":{
                           "title":"Id",
                           "description":"Resource identifier",
                           "type":"string"
                        }
                     }
                  }
               }
            }
         }
      }
   }
}`
)

type RegisterTestService struct {
	testService *testkit.TestService
	state       RegisterTestServiceState
}

type RegisterTestServiceState interface {
	SetServiceClassID(serviceID string)
	GetServiceClassID() string
	GetRegistryClient() *testkit.RegistryClient
}

var _ step.Step = &RegisterTestService{}

func NewRegisterTestService(testService *testkit.TestService, state RegisterTestServiceState) *RegisterTestService {
	return &RegisterTestService{
		testService: testService,
		state:       state,
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
		Provider:         serviceProvider,
		Name:             serviceName,
		Description:      serviceDescription,
		ShortDescription: serviceShortDescription,
		Identifier:       serviceIdentifier,
		Events: &testkit.Events{
			Spec: json.RawMessage(serviceEventsSpec),
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

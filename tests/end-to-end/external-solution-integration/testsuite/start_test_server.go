package testsuite

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
	"time"
)

type StartTestServer struct {
	testService *testkit.TestService
}

var _ step.Step = &StartTestServer{}

func NewStartTestServer(testService *testkit.TestService) *StartTestServer {
	return &StartTestServer{
		testService: testService,
	}
}

func (s *StartTestServer) Name() string {
	return "Start test server"
}

func (s *StartTestServer) Run() error {
	err := s.testService.CreateTestService()

	if err != nil {
		return err
	}

	err = retry.Do(s.testService.IsReady, retry.Delay(200 * time.Millisecond))

	if err != nil {
		return fmt.Errorf("test service not started: %s", err)
	}

	return nil
}

func (s *StartTestServer) Cleanup() error {
	return s.testService.DeleteTestService()
}

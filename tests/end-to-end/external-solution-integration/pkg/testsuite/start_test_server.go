package testsuite

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/pkg/errors"
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

	err = retry.Do(s.testService.IsReady)

	if err != nil {
		return errors.Wrap(err, "test service not started")
	}

	return nil
}

func (s *StartTestServer) Cleanup() error {
	return s.testService.DeleteTestService()
}

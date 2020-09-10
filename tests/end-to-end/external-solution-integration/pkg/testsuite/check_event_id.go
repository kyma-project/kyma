package testsuite

import (
	retrygo "github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

// CheckEventId is a step which checks if the correct EventId has been received
type CheckEventId struct {
	testService *testkit.TestService
	eventId     string
	retryOpts   []retrygo.Option
}

var _ step.Step = &CheckEventId{}

// NewCheckEventId returns new CheckEventId
func NewCheckEventId(testService *testkit.TestService, eventId string, opts ...retrygo.Option) *CheckEventId {
	return &CheckEventId{
		testService: testService,
		eventId:     eventId,
		retryOpts:   opts,
	}
}

// Name returns name name of the step
func (s *CheckEventId) Name() string {
	return "Check event id"
}

// Run executes the step
func (s *CheckEventId) Run() error {
	err := retry.Do(func() error {
		return s.testService.CheckEventId(s.eventId)
	}, s.retryOpts...)
	if err != nil {
		//if all retries fail, the test should ask the subscriber for all events which it received
		return s.testService.CheckAllReceivedEvents()
		//return errors.Wrapf(err, "the event id could not be checked")
	}
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *CheckEventId) Cleanup() error {
	return nil
}

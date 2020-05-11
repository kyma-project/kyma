package event_mesh

import (
	"time"

	retrygo "github.com/avast/retry-go"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
)

const retryAttemptsCount = 240
const retryDelay = 1 * time.Second

var opts = []retrygo.Option{
	retrygo.Attempts(retryAttemptsCount),
	retrygo.Delay(retryDelay),
}

// NewCheckCounterPod returns a new CheckCounterPod
func NewWrappedCounterPod(testService *testkit.TestService, expectedValue int) *testsuite.CheckCounterPod {
	return testsuite.NewCheckCounterPod(testService, expectedValue, opts...)
}

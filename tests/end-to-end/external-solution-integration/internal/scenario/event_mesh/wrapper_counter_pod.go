package event_mesh

import (
	"time"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	"github.com/sirupsen/logrus"
)

const retryAttemptsCount = 120
const retryDelay = 2 * time.Second

var opts = []retry.Option{
	retry.Attempts(retryAttemptsCount),
	retry.Delay(retryDelay),
	retry.OnRetry(func(n uint, err error) {
		logrus.WithField("component", "WrappedCounterPod").Debugf("OnRetry: attempts: %d, error: %v", n, err)
	}),
}

// NewCheckCounterPod returns a new CheckCounterPod
func NewWrappedCounterPod(testService *testkit.TestService, expectedValue int) *testsuite.CheckCounterPod {
	return testsuite.NewCheckCounterPod(testService, expectedValue, opts...)
}

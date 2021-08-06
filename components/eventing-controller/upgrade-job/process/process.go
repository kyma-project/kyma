package process

import (
	"time"

	"github.com/kyma-project/kyma/common/logging/logger"
)

// Process contains the common resources for this upgrade-job steps
type Process struct {
	Steps          []Step
	ReleaseName    string
	KymaNamespace  string
	ControllerName string
	PublisherName  string
	Clients        Clients
	State          State
	TimeoutPeriod  time.Duration
	Logger         *logger.Logger
}

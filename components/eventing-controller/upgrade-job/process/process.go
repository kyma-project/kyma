package process

import (
	"time"

	"github.com/kyma-project/kyma/common/logging/logger"
)

type Process struct {
	Steps          	 []Step
	ReleaseName  	 string
	KymaNamespace 	 string
	ControllerName 	 string
	PublisherName 	 string
	Clients          Clients
	State            State
	TimeoutPeriod 	 time.Duration
	Logger 			 *logger.Logger
}

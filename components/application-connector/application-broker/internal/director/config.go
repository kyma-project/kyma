package director

import "time"

type ServiceConfig struct {
	OperationPollingTimeout  time.Duration `default:"20m"`
	OperationPollingInterval time.Duration `default:"1m"`
}

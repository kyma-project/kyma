package repeat

import (
	"time"
)

// Default constraints
type Config struct {
	Interval time.Duration `envconfig:"default=1s"`
	Timeout  time.Duration `envconfig:"default=5m"`
}

var config = Config{
	Interval: time.Second,
	Timeout:  5 * time.Minute,
}

func SetConfig(cfg Config) {
	config = cfg
}

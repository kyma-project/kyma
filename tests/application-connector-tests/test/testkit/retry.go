package testkit

import (
	"time"
)

type RetryConfig struct {
	MaxRetries int
	Duration   time.Duration
	Factor     float64
}

var DefaultRetryConfig RetryConfig = RetryConfig{
	MaxRetries: 3,
	Duration:   time.Second,
	Factor:     1.5,
}

func RetryOnError(config RetryConfig, function func() error) error {
	var err error

	duration := config.Duration

	for i := 0; i < config.MaxRetries; i++ {
		if i != 0 {
			time.Sleep(duration)
			duration = duration * time.Duration(config.Factor)
		}

		if err = function(); err == nil {
			return nil
		}
	}

	return err
}

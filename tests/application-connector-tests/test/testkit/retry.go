package testkit

import (
	"fmt"
	"time"
)

type RetryConfig struct {
	MaxRetries int
	Duration   time.Duration
	Factor     float64
}

var DefaultRetryConfig RetryConfig = RetryConfig{
	MaxRetries: 5,
	Duration:   2 * time.Second,
	Factor:     1.5,
}

func Retry(config RetryConfig, shouldRetry func() (bool, error)) error {
	duration := config.Duration

	for i := 0; i < config.MaxRetries; i++ {
		if i != 0 {
			time.Sleep(duration)
			duration = duration * time.Duration(config.Factor)
		}

		retry, err := shouldRetry()
		if err != nil {
			return err
		}

		if !retry {
			return nil
		}
	}

	return fmt.Errorf("error: retries limit reachd")
}

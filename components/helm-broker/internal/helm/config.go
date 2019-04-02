package helm

import "time"

// Config holds configuration for helm client
type Config struct {
	TillerHost              string
	TillerConnectionTimeout time.Duration `envconfig:"default=5s"`
}

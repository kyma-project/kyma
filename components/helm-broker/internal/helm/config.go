package helm

import "time"

// Config holds configuration for helm client
type Config struct {
	TillerHost              string
	TillerConnectionTimeout time.Duration `envconfig:"default=5s"`
	TillerTLSKey            string        `envconfig:"default=/etc/certs/tls.key"`
	TillerTLSCrt            string        `envconfig:"default=/etc/certs/tls.crt"`
	TillerTLSInsecure       bool          `envconfig:"default=false"`
}

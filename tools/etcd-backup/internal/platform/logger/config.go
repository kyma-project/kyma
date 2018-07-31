package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type (
	// LogLevel is a config field type holding minimal log level.
	// It's compatible with logrus level types.
	LogLevel logrus.Level

	// Config is responsible for configuring logger.
	Config struct {
		// Level sets the minimal logging level
		// values: debug, info (default), warn, warning, error, fatal, panic
		Level LogLevel `envconfig:"default=info"`

		// BuildHash holds hash of the git commit from which binary was build.
		BuildHash string `envconfig:"-"`
	}
)

// Unmarshal provides custom parsing of Log Level.
// Implements envconfig.Unmarshal interface.
func (m *LogLevel) Unmarshal(in string) error {
	out, err := logrus.ParseLevel(in)

	if err != nil {
		return fmt.Errorf("unable to unmarshal %s", in)
	}

	*m = LogLevel(out)

	return nil
}

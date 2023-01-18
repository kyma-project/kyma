package globalconfig

import (
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/configureLogger"
)

type LogLevelChanger interface {
	ChangeLogLevel(level string) error
	SetDefaultLogLevel() error
}

type OverrideConfig struct {
	LogLevel string `yaml:"logLevel,omitempty"`
}

type ManagerGlobalConfig struct {
	logLevelChanger LogLevelChanger
}

func New(loglevelChanger *configureLogger.LogLevel) *ManagerGlobalConfig {
	var m ManagerGlobalConfig
	m.logLevelChanger = loglevelChanger
	return &m
}

func (m *ManagerGlobalConfig) CheckGlobalConfig(config OverrideConfig) error {
	if config.LogLevel == "" {
		m.logLevelChanger.SetDefaultLogLevel()
	}
	return m.logLevelChanger.ChangeLogLevel(config.LogLevel)
}

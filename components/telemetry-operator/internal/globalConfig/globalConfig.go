package globalConfig

import (
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/configureLogger"
)

type LogLevelChanger interface {
	ReconfigureLogLevel(level map[string]interface{}) error
}

type ManagerGlobalConfig struct {
	logLevelChanger LogLevelChanger
}

func New(loglevelChanger *configureLogger.LogLevel) *ManagerGlobalConfig {
	var m ManagerGlobalConfig
	m.logLevelChanger = loglevelChanger
	return &m
}

func (m *ManagerGlobalConfig) CheckGlobalConfig(config map[string]interface{}) error {
	return m.logLevelChanger.ReconfigureLogLevel(config)
}

package overrides

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/logger"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type LogLevelChanger interface {
	ChangeLogLevel(level string) error
	SetDefaultLogLevel() error
}

type GlobalConfigHandler interface {
	CheckGlobalConfig(config GlobalConfig) error
	UpdateOverrideConfig(ctx context.Context, overrideConfigMap types.NamespacedName) (Config, error)
}

//go:generate mockery --name ConfigMapProber --filename configmap_prober.go
type ConfigMapProber interface {
	ReadConfigMapOrEmpty(ctx context.Context, name types.NamespacedName) (string, error)
}

type Config struct {
	Tracing TracingConfig `yaml:"tracing,omitempty"`
	Logging LoggingConfig `yaml:"logging,omitempty"`
	Global  GlobalConfig  `yaml:"global,omitempty"`
}

type TracingConfig struct {
	Paused bool `yaml:"paused,omitempty"`
}

type LoggingConfig struct {
	Paused bool `yaml:"paused,omitempty"`
}

type GlobalConfig struct {
	LogLevel string `yaml:"logLevel,omitempty"`
}

type Handler struct {
	logLevelChanger LogLevelChanger
	cmProber        ConfigMapProber
}

func New(loglevelChanger *logger.LogLevel, cmProber ConfigMapProber) *Handler {
	var m Handler
	m.logLevelChanger = loglevelChanger
	m.cmProber = cmProber
	return &m
}

func (m *Handler) UpdateOverrideConfig(ctx context.Context, overrideConfigMap types.NamespacedName) (Config, error) {
	log := logf.FromContext(ctx)
	var overrideConfig Config

	config, err := m.cmProber.ReadConfigMapOrEmpty(ctx, overrideConfigMap)
	if err != nil {
		return overrideConfig, err
	}

	if len(config) == 0 {
		return overrideConfig, nil
	}

	err = yaml.Unmarshal([]byte(config), &overrideConfig)
	if err != nil {
		return overrideConfig, err
	}

	log.V(1).Info(fmt.Sprintf("Using override Config is: %+v", overrideConfig))

	return overrideConfig, nil
}

func (m *Handler) CheckGlobalConfig(config GlobalConfig) error {
	if config.LogLevel == "" {
		m.logLevelChanger.SetDefaultLogLevel()
	}

	return m.logLevelChanger.ChangeLogLevel(config.LogLevel)
}

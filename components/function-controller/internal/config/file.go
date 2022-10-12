package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/internal/file"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	notificationDelay = 1 * time.Second
)

type CallbackFn func(Config)

type Config struct {
	LogLevel  string `yaml:"logLevel"`
	LogFormat string `yaml:"logFormat"`
}

// Load - return cfg struct based on given path
func Load(path string) (Config, error) {
	cfg := Config{}

	cleanPath := filepath.Clean(path)
	yamlFile, err := os.ReadFile(cleanPath)
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	return cfg, err
}

// RunOnConfigChange - run callback functions when config is changed
func RunOnConfigChange(ctx context.Context, log *zap.SugaredLogger, path string, callbacks ...CallbackFn) {
	log.Info("config notifier started")

	for {
		// wait 1 sec not to burn out the container for example when any method below always ends with an error
		time.Sleep(notificationDelay)

		err := fireCallbacksOnConfigChange(ctx, log, path, callbacks...)
		if err != nil && errors.Is(err, context.Canceled) {
			log.Info("context canceled")
			return
		}
		if err != nil {
			log.Error(err)
		}
	}
}

func fireCallbacksOnConfigChange(ctx context.Context, log *zap.SugaredLogger, path string, callbacks ...CallbackFn) error {
	err := file.NotifyModification(ctx, path)
	if err != nil {
		return err
	}

	log.Info("config file change detected")

	cfg, err := Load(path)
	if err != nil {
		return err
	}

	log.Debugf("firing '%d' callbacks", len(callbacks))

	fireCallbacks(cfg, callbacks...)
	return nil
}

func fireCallbacks(cfg Config, funcs ...CallbackFn) {
	for i := range funcs {
		fn := funcs[i]
		fn(cfg)
	}
}

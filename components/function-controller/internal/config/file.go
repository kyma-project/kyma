package config

import (
	"context"
	"errors"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/internal/file"
	"go.uber.org/zap"
)

const (
	notificationDelay = 1 * time.Second
)

type CallbackFn func(Config)

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

	cfg, err := LoadLogConfig(path)
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

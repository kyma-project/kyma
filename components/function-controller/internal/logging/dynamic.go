package logging

import (
	"context"

	"github.com/kyma-project/kyma/components/function-controller/internal/config"
)

func ReconfigureOnConfigChange(ctx context.Context, registry *Registry, cfgPath string) {
	notifyLog := registry.CreateNamed("notifier")

	config.RunOnConfigChange(ctx, notifyLog, cfgPath, func(cfg config.Config) {
		err := registry.Reconfigure(cfg.LogLevel, cfg.LogFormat)
		if err != nil {
			notifyLog.Error(err)
			return
		}

		notifyLog.Infof("loggers reconfigured with level '%s' and format '%s'", cfg.LogLevel, cfg.LogFormat)
	})
}

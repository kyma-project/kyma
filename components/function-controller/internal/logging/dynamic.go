package logging

import (
	"context"

	"github.com/kyma-project/kyma/components/function-controller/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ReconfigureOnConfigChange(ctx context.Context, log *zap.SugaredLogger, atomic zap.AtomicLevel, cfgPath string) {
	config.RunOnConfigChange(ctx, log, cfgPath, func(cfg config.Config) {
		level, err := zapcore.ParseLevel(cfg.LogLevel)
		if err != nil {
			log.Error(err)
			return
		}

		atomic.SetLevel(level)
		log.Infof("loggers reconfigured with level '%s' and format '%s'", cfg.LogLevel, cfg.LogFormat)
	})
}

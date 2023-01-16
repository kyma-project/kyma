package logpipeline

import (
	"context"
	"github.com/imdario/mergo"
)

func (r *Reconciler) UpdateOverrideConfig(ctx context.Context) (map[string]interface{}, error) {
	overrideConfig := make(map[string]interface{})

	config, err := r.cmProber.IsPresent(ctx, r.config.OverrideConfigMap)
	if err != nil {
		return overrideConfig, err
	}

	if len(config) == 0 {
		return overrideConfig, nil
	}

	traceConfig := fetchTracingConfig(config)
	if err := mergo.Merge(&overrideConfig, &traceConfig); err != nil {
		return overrideConfig, err
	}
	globalConfig := fetchGlobalConfig(config)
	if err = mergo.Merge(&overrideConfig, &globalConfig); err != nil {
		return overrideConfig, err
	}

	return overrideConfig, nil
}

func fetchTracingConfig(config map[string]interface{}) map[string]interface{} {
	logConfig := make(map[string]interface{})
	overrideConfig := make(map[string]interface{})
	if _, ok := config["logging"]; !ok {
		return overrideConfig
	}
	logConfig = config["logging"].(map[string]interface{})
	if paused, ok := logConfig["paused"]; ok {
		overrideConfig["paused"] = paused
	}
	return overrideConfig
}

func fetchGlobalConfig(config map[string]interface{}) map[string]interface{} {
	globalConfig := make(map[string]interface{})
	overrideConfig := make(map[string]interface{})

	if _, ok := config["global"]; !ok {
		return overrideConfig
	}
	globalConfig = config["global"].(map[string]interface{})

	if logLevel, ok := globalConfig["logLevel"]; ok {
		overrideConfig["logLevel"] = logLevel
	}
	return overrideConfig
}

func (r *Reconciler) pauseReconciliation(overrideConfig map[string]interface{}) bool {
	if paused, ok := overrideConfig["paused"]; ok {
		return paused.(bool)
	}
	return false
}

func (r *Reconciler) reconfigureLogLevel(overrideConfig map[string]interface{}) error {
	if logLevel, ok := overrideConfig["logLevel"].(string); ok {
		return r.logLevel.ChangeLogLevel(logLevel)
	}
	return nil
}

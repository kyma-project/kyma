package logpipeline

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *Reconciler) UpdateOverrideConfig(ctx context.Context) (OverrideConfig, error) {
	log := logf.FromContext(ctx)
	var overrideConfig OverrideConfig

	config, err := r.cmProber.IsPresent(ctx, r.config.OverrideConfigMap)
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

	log.V(1).Info(fmt.Sprintf("Override Config is: %v", overrideConfig))

	return overrideConfig, nil
}

func (r *Reconciler) pauseReconciliation(overrideConfig LoggingConfig) bool {
	return overrideConfig.Paused
}

package logparser

import (
	"context"
	"fmt"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
	utils "github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const parsersConfigMapKey = "parsers.conf"

type syncer struct {
	client.Client
	config Config
}

func (s *syncer) syncFluentBitConfig(ctx context.Context) error {
	cm, err := utils.GetOrCreateConfigMap(ctx, s, s.config.ParsersConfigMap)
	if err != nil {
		return fmt.Errorf("unable to get parsers configmap: %w", err)
	}

	changed := false
	var logParsers telemetryv1alpha1.LogParserList

	err = s.List(ctx, &logParsers)
	if err != nil {
		return fmt.Errorf("unable to list parsers: %w", err)
	}
	fluentBitParsersConfig := builder.BuildFluentBitParsersConfig(&logParsers)
	if fluentBitParsersConfig == "" {
		data := make(map[string]string)
		data[parsersConfigMapKey] = ""
		cm.Data = data
		changed = true
	} else if cm.Data == nil {
		data := make(map[string]string)
		data[parsersConfigMapKey] = fluentBitParsersConfig
		cm.Data = data
		changed = true
	} else {
		if oldConfig, hasKey := cm.Data[parsersConfigMapKey]; !hasKey || oldConfig != fluentBitParsersConfig {
			cm.Data[parsersConfigMapKey] = fluentBitParsersConfig
			changed = true
		}
	}

	if !changed {
		return nil
	}

	if err = s.Update(ctx, &cm); err != nil {
		return fmt.Errorf("unable to parsers files configmap: %w", err)
	}

	return nil
}

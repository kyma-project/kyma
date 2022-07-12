package parsers

import (
	"context"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	parsersConfigMapKey      = "parsers.conf"
	parserConfigMapFinalizer = "FLUENT_BIT_PARSERS_CONFIG_MAP"
)

type FluentBitDaemonSetConfig struct {
	FluentBitDaemonSetName    types.NamespacedName
	FluentBitParsersConfigMap types.NamespacedName
}
type LogParserSyncer struct {
	client.Client
	DaemonSetConfig FluentBitDaemonSetConfig
	Utils           *kubernetes.Utils
}

func NewLogParserSyncer(client client.Client,
	daemonSetConfig FluentBitDaemonSetConfig,
) *LogParserSyncer {
	var lps LogParserSyncer
	lps.Client = client
	lps.DaemonSetConfig = daemonSetConfig
	lps.Utils = kubernetes.NewUtils(client)
	return &lps
}

// Synchronize LogPipeline with ConfigMap of FluentBitUtils parsers (Parser and MultiLineParser).
func (s *LogParserSyncer) SyncParsersConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogParser) (bool, error) {
	log := logf.FromContext(ctx)
	cm, err := s.Utils.GetOrCreateConfigMap(ctx, s.DaemonSetConfig.FluentBitParsersConfigMap)
	if err != nil {
		return false, err
	}

	changed := false
	var logParser telemetryv1alpha1.LogParserList

	if logPipeline.DeletionTimestamp != nil {
		if cm.Data != nil && controllerutil.ContainsFinalizer(logPipeline, parserConfigMapFinalizer) {
			log.Info("Deleting fluent bit parsers config")

			err = s.List(ctx, &logParser)
			if err != nil {
				return false, err
			}

			fluentBitParsersConfig := fluentbit.MergeParsersConfig(&logParser)
			if fluentBitParsersConfig == "" {
				cm.Data = nil
			} else {
				data := make(map[string]string)
				data[parsersConfigMapKey] = fluentBitParsersConfig
				cm.Data = data
			}
			controllerutil.RemoveFinalizer(logPipeline, parserConfigMapFinalizer)
			changed = true
		}
	} else {
		err = s.List(ctx, &logParser)
		if err != nil {
			return false, err
		}

		fluentBitParsersConfig := fluentbit.MergeParsersConfig(&logParser)
		if fluentBitParsersConfig == "" {
			if cm.Data == nil {
				return false, nil
			}
			cm.Data = nil
		} else {

			if cm.Data == nil {
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
			if !controllerutil.ContainsFinalizer(logPipeline, parserConfigMapFinalizer) {
				log.Info("Adding finalizer")
				controllerutil.AddFinalizer(logPipeline, parserConfigMapFinalizer)
				changed = true
			}
		}
	}

	if !changed {
		return false, nil
	}
	if err = s.Update(ctx, &cm); err != nil {
		return false, err
	}

	return changed, nil
}

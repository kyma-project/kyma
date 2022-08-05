package sync

import (
	"context"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/controller/logpipeline/fluentbitconfig"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
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

type Config struct {
	FluentBitDaemonSetName    types.NamespacedName
	FluentBitParsersConfigMap types.NamespacedName
}
type Syncer struct {
	client.Client
	DaemonSetConfig Config
	Utils           *kubernetes.Utils
}

func NewSyncer(client client.Client, config Config) *Syncer {
	var lps Syncer
	lps.Client = client
	lps.DaemonSetConfig = config
	lps.Utils = kubernetes.NewUtils(client)
	return &lps
}

// SyncParsersConfigMap synchronizes LogParser with ConfigMap that contains Fluent Bit parsers
func (s *Syncer) SyncParsersConfigMap(ctx context.Context, logParser *telemetryv1alpha1.LogParser) (bool, error) {
	log := logf.FromContext(ctx)
	cm, err := s.Utils.GetOrCreateConfigMap(ctx, s.DaemonSetConfig.FluentBitParsersConfigMap)
	if err != nil {
		return false, err
	}

	changed := false
	var logParsers telemetryv1alpha1.LogParserList

	if logParser.DeletionTimestamp != nil {
		if cm.Data != nil && controllerutil.ContainsFinalizer(logParser, parserConfigMapFinalizer) {
			log.Info("Deleting fluent bit parsers config")

			err = s.List(ctx, &logParsers)
			if err != nil {
				return false, err
			}

			fluentBitParsersConfig := fluentbitconfig.MergeParsersConfig(&logParsers)
			if fluentBitParsersConfig == "" {
				cm.Data = nil
			} else {
				data := make(map[string]string)
				data[parsersConfigMapKey] = fluentBitParsersConfig
				cm.Data = data
			}
			controllerutil.RemoveFinalizer(logParser, parserConfigMapFinalizer)
			changed = true
		}
	} else {
		err = s.List(ctx, &logParsers)
		if err != nil {
			return false, err
		}

		fluentBitParsersConfig := fluentbitconfig.MergeParsersConfig(&logParsers)
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
		if !controllerutil.ContainsFinalizer(logParser, parserConfigMapFinalizer) {
			log.Info("Adding finalizer")
			controllerutil.AddFinalizer(logParser, parserConfigMapFinalizer)
			changed = true
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

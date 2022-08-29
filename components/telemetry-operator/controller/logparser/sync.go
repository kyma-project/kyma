package logparser

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
)

const (
	parsersConfigMapKey      = "parsers.conf"
	parserConfigMapFinalizer = "FLUENT_BIT_PARSERS_CONFIG_MAP"
)

type syncer struct {
	client.Client
	fluentBitK8sResources fluentbit.KubernetesResources
	k8sGetterOrCreator    *kubernetes.GetterOrCreator
}

func newSyncer(client client.Client, fluentBitK8sResources fluentbit.KubernetesResources) *syncer {
	var s syncer
	s.Client = client
	s.fluentBitK8sResources = fluentBitK8sResources
	s.k8sGetterOrCreator = kubernetes.NewGetterOrCreator(client)
	return &s
}

// SyncParsersConfigMap synchronizes the Fluent Bit parsers ConfigMap for all LogParsers.
func (s *syncer) SyncParsersConfigMap(ctx context.Context, logParser *telemetryv1alpha1.LogParser) (bool, error) {
	log := logf.FromContext(ctx)
	cm, err := s.k8sGetterOrCreator.ConfigMap(ctx, s.fluentBitK8sResources.ParsersConfigMap)
	if err != nil {
		return false, err
	}

	changed := false
	var logParsers telemetryv1alpha1.LogParserList

	if logParser.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(logParser, parserConfigMapFinalizer) {
			log.Info("Adding finalizer")
			controllerutil.AddFinalizer(logParser, parserConfigMapFinalizer)
			changed = true
		}
	} else {
		if controllerutil.ContainsFinalizer(logParser, parserConfigMapFinalizer) {
			log.Info("Removing finalizer")
			controllerutil.RemoveFinalizer(logParser, parserConfigMapFinalizer)
			changed = true
		}
	}

	err = s.List(ctx, &logParsers)
	if err != nil {
		return false, err
	}
	fluentBitParsersConfig := builder.BuildFluentBitParsersConfig(&logParsers)
	if fluentBitParsersConfig == "" {
		cm.Data = nil
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
		return false, nil
	}
	if err = s.Update(ctx, &cm); err != nil {
		return false, err
	}

	return changed, nil
}

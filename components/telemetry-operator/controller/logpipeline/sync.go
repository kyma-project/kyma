package logpipeline

import (
	"context"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

const (
	sectionsConfigMapFinalizer = "FLUENT_BIT_SECTIONS_CONFIG_MAP"
	filesFinalizer             = "FLUENT_BIT_FILES"
)

type syncer struct {
	client.Client
	config                  Config
	enableUnsupportedPlugin bool
	unsupportedPluginsTotal int
	secretHelper            *secretHelper
	k8sGetterOrCreator      *kubernetes.GetterOrCreator
}

func newSyncer(
	client client.Client,
	config Config,
) *syncer {
	var s syncer
	s.Client = client
	s.config = config
	s.secretHelper = newSecretHelper(client)
	s.k8sGetterOrCreator = kubernetes.NewGetterOrCreator(client)
	return &s
}

func (s *syncer) SyncAll(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := logf.FromContext(ctx)

	sectionsChanged, err := s.syncSectionsConfigMap(ctx, logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync Sections ConfigMap")
		return false, err
	}

	filesChanged, err := s.syncFilesConfigMap(ctx, logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync mounted files")
		return false, err
	}

	var logPipelines telemetryv1alpha1.LogPipelineList
	err = s.List(ctx, &logPipelines)
	if err != nil {
		return false, err
	}

	variablesChanged, err := s.syncVariables(ctx, &logPipelines)
	if err != nil {
		log.Error(err, "Failed to sync variables")
		return false, err
	}

	s.syncUnsupportedPluginsTotal(&logPipelines)

	return sectionsChanged || filesChanged || variablesChanged, nil
}

// Synchronize LogPipeline with ConfigMap of daemonSetHelper sections (Input, Filter and Output).
func (s *syncer) syncSectionsConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := logf.FromContext(ctx)
	cm, err := s.k8sGetterOrCreator.ConfigMap(ctx, s.config.SectionsConfigMap)
	if err != nil {
		return false, err
	}

	changed := false
	cmKey := logPipeline.Name + ".conf"
	if logPipeline.DeletionTimestamp != nil {
		if cm.Data != nil && controllerutil.ContainsFinalizer(logPipeline, sectionsConfigMapFinalizer) {
			log.Info("Deleting fluent bit config")
			delete(cm.Data, cmKey)
			controllerutil.RemoveFinalizer(logPipeline, sectionsConfigMapFinalizer)
			changed = true
		}
	} else {
		fluentBitConfig, err := builder.BuildFluentBitConfig(logPipeline, s.config.PipelineDefaults)
		if err != nil {
			return false, err
		}
		if cm.Data == nil {
			data := make(map[string]string)
			data[cmKey] = fluentBitConfig
			cm.Data = data
			changed = true
		} else if oldConfig, hasKey := cm.Data[cmKey]; !hasKey || oldConfig != fluentBitConfig {
			cm.Data[cmKey] = fluentBitConfig
			changed = true
		}
		if !controllerutil.ContainsFinalizer(logPipeline, sectionsConfigMapFinalizer) {
			log.Info("Adding finalizer")
			controllerutil.AddFinalizer(logPipeline, sectionsConfigMapFinalizer)
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

// Synchronize file references with Fluent Bit files ConfigMap.
func (s *syncer) syncFilesConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := logf.FromContext(ctx)
	cm, err := s.k8sGetterOrCreator.ConfigMap(ctx, s.config.FilesConfigMap)
	if err != nil {
		return false, err
	}

	changed := false
	for _, file := range logPipeline.Spec.Files {
		if logPipeline.DeletionTimestamp != nil {
			if _, hasKey := cm.Data[file.Name]; hasKey {
				delete(cm.Data, file.Name)
				controllerutil.RemoveFinalizer(logPipeline, filesFinalizer)
				changed = true
			}
		} else {
			if cm.Data == nil {
				data := make(map[string]string)
				data[file.Name] = file.Content
				cm.Data = data
				changed = true
			} else if oldContent, hasKey := cm.Data[file.Name]; !hasKey || oldContent != file.Content {
				cm.Data[file.Name] = file.Content
				changed = true
			}
			if !controllerutil.ContainsFinalizer(logPipeline, filesFinalizer) {
				log.Info("Adding finalizer")
				controllerutil.AddFinalizer(logPipeline, filesFinalizer)
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

// syncVariables copies referenced secrets to global Fluent Bit environment secret.
func (s *syncer) syncVariables(ctx context.Context, logPipelines *telemetryv1alpha1.LogPipelineList) (bool, error) {
	log := logf.FromContext(ctx)
	oldSecret, err := s.k8sGetterOrCreator.Secret(ctx, s.config.EnvSecret)
	if err != nil {
		return false, err
	}

	newSecret := oldSecret
	newSecret.Data = make(map[string][]byte)

	for _, l := range logPipelines.Items {
		for _, field := range lookupSecretRefFields(&l) {
			err := s.secretHelper.CopySecretData(ctx, *&field.secretKeyRef, field.targetSecretKey, newSecret.Data)
			if err != nil {
				log.Error(err, "unable to find secret for http host")
				return false, err
			}
		}
	}

	secretHasChanged := SecretHasChanged(oldSecret.Data, newSecret.Data)
	if !secretHasChanged {
		return false, nil
	}

	if err = s.Update(ctx, &newSecret); err != nil {
		log.Error(err, err.Error())
		return false, err
	}
	return secretHasChanged, nil
}

// syncUnsupportedPluginsTotal checks if any LogPipeline defines a unsupported Filter or Output.
func (s *syncer) syncUnsupportedPluginsTotal(logPipelines *telemetryv1alpha1.LogPipelineList) {
	unsupportedPluginsTotal := 0
	for _, l := range logPipelines.Items {
		if !l.DeletionTimestamp.IsZero() {
			continue
		}
		if IsUnsupported(l) {
			unsupportedPluginsTotal++
		}
	}

	s.unsupportedPluginsTotal = unsupportedPluginsTotal
}

func IsUnsupported(pipeline telemetryv1alpha1.LogPipeline) bool {
	if pipeline.Spec.Output.Custom != "" {
		return true
	}
	for _, f := range pipeline.Spec.Filters {
		if f.Custom != "" {
			return true
		}
	}
	return false
}

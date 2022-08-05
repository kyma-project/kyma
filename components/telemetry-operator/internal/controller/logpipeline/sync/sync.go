package sync

import (
	"context"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/controller/logpipeline/fluentbitconfig"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	sectionsConfigMapFinalizer = "FLUENT_BIT_SECTIONS_CONFIG_MAP"
	filesFinalizer             = "FLUENT_BIT_FILES"
)

type Config struct {
	FluentBitDaemonSetName     types.NamespacedName
	FluentBitSectionsConfigMap types.NamespacedName
	FluentBitFilesConfigMap    types.NamespacedName
	FluentBitEnvSecret         types.NamespacedName
}

type Syncer struct {
	client.Client
	DaemonSetConfig         Config
	PipelineConfig          fluentbitconfig.PipelineConfig
	EnableUnsupportedPlugin bool
	UnsupportedPluginsTotal int
	SecretHelper            *Helper
	Utils                   *kubernetes.Utils
}

func NewSyncer(client client.Client, config Config, pipelineConfig fluentbitconfig.PipelineConfig) *Syncer {
	var syncer Syncer
	syncer.Client = client
	syncer.DaemonSetConfig = config
	syncer.PipelineConfig = pipelineConfig
	syncer.SecretHelper = NewSecretHelper(client)
	syncer.Utils = kubernetes.NewUtils(client)
	return &syncer
}

func (s *Syncer) SyncAll(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
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
	variablesChanged, err := s.syncVariables(ctx)
	if err != nil {
		log.Error(err, "Failed to sync variables")
		return false, err
	}

	err = s.syncUnsupportedPluginsTotal(ctx)
	if err != nil {
		log.Error(err, "Failed to sync unsupported mode metrics")
		return false, err
	}

	return sectionsChanged || filesChanged || variablesChanged, nil
}

// Synchronize LogPipeline with ConfigMap of DaemonSetUtils sections (Input, Filter and Output).
func (s *Syncer) syncSectionsConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := logf.FromContext(ctx)
	cm, err := s.Utils.GetOrCreateConfigMap(ctx, s.DaemonSetConfig.FluentBitSectionsConfigMap)
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
		fluentBitConfig, err := fluentbitconfig.MergeSectionsConfig(logPipeline, s.PipelineConfig)
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

func (s *Syncer) syncUnsupportedPluginsTotal(ctx context.Context) error {
	var logPipelines telemetryv1alpha1.LogPipelineList
	err := s.List(ctx, &logPipelines)
	if err != nil {
		return err
	}

	s.UnsupportedPluginsTotal = updateUnsupportedPluginsTotal(&logPipelines)
	return nil
}

// Synchronize file references with Fluent Bit files ConfigMap.
func (s *Syncer) syncFilesConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := logf.FromContext(ctx)
	cm, err := s.Utils.GetOrCreateConfigMap(ctx, s.DaemonSetConfig.FluentBitFilesConfigMap)
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
func (s *Syncer) syncVariables(ctx context.Context) (bool, error) {
	log := logf.FromContext(ctx)
	oldSecret, err := s.Utils.GetOrCreateSecret(ctx, s.DaemonSetConfig.FluentBitEnvSecret)
	if err != nil {
		return false, err
	}

	newSecret := oldSecret
	newSecret.Data = make(map[string][]byte)

	var logPipelines telemetryv1alpha1.LogPipelineList
	err = s.List(ctx, &logPipelines)
	if err != nil {
		return false, err
	}

	for _, l := range logPipelines.Items {
		if l.DeletionTimestamp != nil {
			continue
		}
		for _, varRef := range l.Spec.Variables {
			if varRef.ValueFrom.IsSecretRef() {
				err := s.SecretHelper.copySecretData(ctx, varRef.ValueFrom, varRef.Name, newSecret.Data)
				if err != nil {
					log.Error(err, "unable to find secret for environment variable")
					return false, err
				}
			}
		}
		if l.Spec.Output.HTTP.Host.ValueFrom.IsSecretRef() {
			err := s.SecretHelper.copySecretData(ctx, l.Spec.Output.HTTP.Host.ValueFrom, l.Spec.Output.HTTP.Host.ValueFrom.SecretKey.EnvVarName(l.Name), newSecret.Data)
			if err != nil {
				log.Error(err, "unable to find secret for http host")
				return false, err
			}
		}
		if l.Spec.Output.HTTP.User.ValueFrom.IsSecretRef() {
			err := s.SecretHelper.copySecretData(ctx, l.Spec.Output.HTTP.User.ValueFrom, l.Spec.Output.HTTP.User.ValueFrom.SecretKey.EnvVarName(l.Name), newSecret.Data)
			if err != nil {
				log.Error(err, "unable to find secret for http user")
				return false, err
			}
		}
		if l.Spec.Output.HTTP.Password.ValueFrom.IsSecretRef() {
			err := s.SecretHelper.copySecretData(ctx, l.Spec.Output.HTTP.Password.ValueFrom, l.Spec.Output.HTTP.Password.ValueFrom.SecretKey.EnvVarName(l.Name), newSecret.Data)
			if err != nil {
				log.Error(err, "unable to find secret for http password")
				return false, err
			}
		}
	}

	needsSecretUpdate := checkIfSecretHasChanged(newSecret.Data, oldSecret.Data)
	if !needsSecretUpdate {
		return false, nil
	}

	if err = s.Update(ctx, &newSecret); err != nil {
		log.Error(err, err.Error())
		return false, err
	}
	return needsSecretUpdate, nil
}

func updateUnsupportedPluginsTotal(pipelines *telemetryv1alpha1.LogPipelineList) int {
	unsupportedPluginsTotal := 0
	for _, l := range pipelines.Items {
		if l.DeletionTimestamp != nil {
			continue
		}
		if LogPipelineIsUnsupported(l) {
			unsupportedPluginsTotal++
		}

	}
	return unsupportedPluginsTotal
}

func LogPipelineIsUnsupported(pipeline telemetryv1alpha1.LogPipeline) bool {
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

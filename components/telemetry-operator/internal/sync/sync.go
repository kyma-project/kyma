package sync

import (
	"context"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/secret"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	sectionsConfigMapFinalizer = "FLUENT_BIT_SECTIONS_CONFIG_MAP"
	filesFinalizer             = "FLUENT_BIT_FILES"
)

type FluentBitDaemonSetConfig struct {
	FluentBitDaemonSetName     types.NamespacedName
	FluentBitSectionsConfigMap types.NamespacedName
	FluentBitFilesConfigMap    types.NamespacedName
	FluentBitEnvSecret         types.NamespacedName
}

type LogPipelineSyncer struct {
	client.Client
	DaemonSetConfig         FluentBitDaemonSetConfig
	PipelineConfig          fluentbit.PipelineConfig
	EnableUnsupportedPlugin bool
	UnsupportedPluginsTotal int
	SecretHelper            *secret.Helper
	Utils                   *kubernetes.Utils
}

type Result struct {
	ConfigurationChanged bool
	LogPipelineChanged   bool
}

func NewLogPipelineSyncer(client client.Client,
	daemonSetConfig FluentBitDaemonSetConfig,
	pipelineConfig fluentbit.PipelineConfig,
) *LogPipelineSyncer {
	var lps LogPipelineSyncer
	lps.Client = client
	lps.DaemonSetConfig = daemonSetConfig
	lps.PipelineConfig = pipelineConfig
	lps.SecretHelper = secret.NewSecretHelper(client)
	lps.Utils = kubernetes.NewUtils(client)
	return &lps
}

func (s *LogPipelineSyncer) SyncAll(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (Result, error) {
	var syncRes Result
	log := logf.FromContext(ctx)

	sectionsChanged, err := s.syncSectionsConfigMap(ctx, logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync Sections ConfigMap")
		return syncRes, err
	}

	filesChanged, err := s.syncFilesConfigMap(ctx, logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync mounted files")
		return syncRes, err
	}
	variablesChanged, err := s.syncVariables(ctx)
	if err != nil {
		log.Error(err, "Failed to sync variables")
		return syncRes, err
	}

	err = s.syncUnsupportedPluginsTotal(ctx)
	if err != nil {
		log.Error(err, "Failed to sync unsupported mode metrics")
		return syncRes, err
	}

	if sectionsChanged.ConfigurationChanged || filesChanged.ConfigurationChanged || variablesChanged.ConfigurationChanged {
		syncRes.ConfigurationChanged = true
	} else if sectionsChanged.LogPipelineChanged || filesChanged.LogPipelineChanged || variablesChanged.LogPipelineChanged {
		syncRes.LogPipelineChanged = true
	}

	return syncRes, nil
}

// Synchronize LogPipeline with ConfigMap of DaemonSetUtils sections (Input, Filter and Output).
func (s *LogPipelineSyncer) syncSectionsConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (Result, error) {
	log := logf.FromContext(ctx)
	var syncRes Result
	cm, err := s.Utils.GetOrCreateConfigMap(ctx, s.DaemonSetConfig.FluentBitSectionsConfigMap)
	if err != nil {
		return syncRes, err
	}

	cmKey := logPipeline.Name + ".conf"
	if logPipeline.DeletionTimestamp != nil {
		if cm.Data != nil && controllerutil.ContainsFinalizer(logPipeline, sectionsConfigMapFinalizer) {
			log.Info("Deleting fluent bit config")
			delete(cm.Data, cmKey)
			controllerutil.RemoveFinalizer(logPipeline, sectionsConfigMapFinalizer)
			syncRes.LogPipelineChanged = true
			syncRes.ConfigurationChanged = true
		}
	} else {
		fluentBitConfig, err := fluentbit.MergeSectionsConfig(logPipeline, s.PipelineConfig)
		if err != nil {
			return syncRes, err
		}
		if cm.Data == nil {
			data := make(map[string]string)
			data[cmKey] = fluentBitConfig
			cm.Data = data
			syncRes.ConfigurationChanged = true
		} else if oldConfig, hasKey := cm.Data[cmKey]; !hasKey || oldConfig != fluentBitConfig {
			cm.Data[cmKey] = fluentBitConfig
			syncRes.ConfigurationChanged = true
		}
		if !controllerutil.ContainsFinalizer(logPipeline, sectionsConfigMapFinalizer) {
			log.Info("Adding finalizer")
			controllerutil.AddFinalizer(logPipeline, sectionsConfigMapFinalizer)
			syncRes.LogPipelineChanged = true
		}
	}

	if !syncRes.LogPipelineChanged && !syncRes.ConfigurationChanged {
		return syncRes, nil
	}
	if err = s.Update(ctx, &cm); err != nil {
		syncRes.LogPipelineChanged = false
		syncRes.ConfigurationChanged = false
		return syncRes, err
	}

	return syncRes, nil
}

func (s *LogPipelineSyncer) syncUnsupportedPluginsTotal(ctx context.Context) error {
	var logPipelines telemetryv1alpha1.LogPipelineList
	err := s.List(ctx, &logPipelines)
	if err != nil {
		return err
	}

	s.UnsupportedPluginsTotal = updateUnsupportedPluginsTotal(&logPipelines)
	return nil
}

// Synchronize file references with Fluent Bit files ConfigMap.
func (s *LogPipelineSyncer) syncFilesConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (Result, error) {
	log := logf.FromContext(ctx)
	var syncRes Result
	cm, err := s.Utils.GetOrCreateConfigMap(ctx, s.DaemonSetConfig.FluentBitFilesConfigMap)
	if err != nil {
		return syncRes, err
	}

	for _, file := range logPipeline.Spec.Files {
		if logPipeline.DeletionTimestamp != nil {
			if _, hasKey := cm.Data[file.Name]; hasKey {
				delete(cm.Data, file.Name)
				controllerutil.RemoveFinalizer(logPipeline, filesFinalizer)
				syncRes.LogPipelineChanged = true
				syncRes.ConfigurationChanged = true
			}
		} else {
			if cm.Data == nil {
				data := make(map[string]string)
				data[file.Name] = file.Content
				cm.Data = data
				syncRes.ConfigurationChanged = true
			} else if oldContent, hasKey := cm.Data[file.Name]; !hasKey || oldContent != file.Content {
				cm.Data[file.Name] = file.Content
				syncRes.ConfigurationChanged = true
			}
			if !controllerutil.ContainsFinalizer(logPipeline, filesFinalizer) {
				log.Info("Adding finalizer")
				controllerutil.AddFinalizer(logPipeline, filesFinalizer)
				syncRes.LogPipelineChanged = true
			}
		}
	}

	if !syncRes.LogPipelineChanged && !syncRes.ConfigurationChanged {
		return syncRes, nil
	}
	if err = s.Update(ctx, &cm); err != nil {
		syncRes.LogPipelineChanged = false
		syncRes.ConfigurationChanged = false
		return syncRes, err
	}

	return syncRes, nil
}

// syncVariables copies referenced secrets to global Fluent Bit environment secret.
func (s *LogPipelineSyncer) syncVariables(ctx context.Context) (Result, error) {
	log := logf.FromContext(ctx)
	var syncRes Result
	oldSecret, err := s.Utils.GetOrCreateSecret(ctx, s.DaemonSetConfig.FluentBitEnvSecret)
	if err != nil {
		return syncRes, err
	}

	newSecret := oldSecret
	newSecret.Data = make(map[string][]byte)

	var logPipelines telemetryv1alpha1.LogPipelineList
	err = s.List(ctx, &logPipelines)
	if err != nil {
		return syncRes, err
	}

	for _, l := range logPipelines.Items {
		if l.DeletionTimestamp != nil {
			continue
		}
		for _, varRef := range l.Spec.Variables {
			if varRef.ValueFrom.IsSecretRef() {
				err := s.SecretHelper.CopySecretData(ctx, varRef.ValueFrom, varRef.Name, newSecret.Data)
				if err != nil {
					log.Error(err, "unable to find secret for environment variable")
					return syncRes, err
				}
			}
		}
		if l.Spec.Output.HTTP.Host.ValueFrom.IsSecretRef() {
			err := s.SecretHelper.CopySecretData(ctx, l.Spec.Output.HTTP.Host.ValueFrom, secret.GenerateVariableName(l.Spec.Output.HTTP.Host.ValueFrom.SecretKey, l.Name), newSecret.Data)
			if err != nil {
				log.Error(err, "unable to find secret for http host")
				return syncRes, err
			}
		}
		if l.Spec.Output.HTTP.User.ValueFrom.IsSecretRef() {
			err := s.SecretHelper.CopySecretData(ctx, l.Spec.Output.HTTP.User.ValueFrom, secret.GenerateVariableName(l.Spec.Output.HTTP.User.ValueFrom.SecretKey, l.Name), newSecret.Data)
			if err != nil {
				log.Error(err, "unable to find secret for http user")
				return syncRes, err
			}
		}
		if l.Spec.Output.HTTP.Password.ValueFrom.IsSecretRef() {
			err := s.SecretHelper.CopySecretData(ctx, l.Spec.Output.HTTP.Password.ValueFrom, secret.GenerateVariableName(l.Spec.Output.HTTP.Password.ValueFrom.SecretKey, l.Name), newSecret.Data)
			if err != nil {
				log.Error(err, "unable to find secret for http password")
				return syncRes, err
			}
		}
	}

	needsSecretUpdate := secret.CheckIfSecretHasChanged(newSecret.Data, oldSecret.Data)
	if !needsSecretUpdate {
		return syncRes, nil
	}
	syncRes.ConfigurationChanged = needsSecretUpdate

	if err = s.Update(ctx, &newSecret); err != nil {
		log.Error(err, err.Error())
		return syncRes, err
	}
	return syncRes, nil
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

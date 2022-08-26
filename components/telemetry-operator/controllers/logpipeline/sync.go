package logpipeline

import (
	"context"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/utils/envvar"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	sectionsConfigMapFinalizer = "FLUENT_BIT_SECTIONS_CONFIG_MAP"
	filesFinalizer             = "FLUENT_BIT_FILES"
)

type syncer struct {
	client.Client
	fluentBitK8sResources   fluentbit.KubernetesResources
	pipelineConfig          builder.PipelineConfig
	enableUnsupportedPlugin bool
	unsupportedPluginsTotal int
	secretHelper            *secretHelper
	k8sGetterOrCreator      *kubernetes.GetterOrCreator
}

func newLogPipelineSyncer(
	client client.Client,
	fluentBitK8sResources fluentbit.KubernetesResources,
	pipelineConfig builder.PipelineConfig,
) *syncer {
	var lps syncer
	lps.Client = client
	lps.fluentBitK8sResources = fluentBitK8sResources
	lps.pipelineConfig = pipelineConfig
	lps.secretHelper = newSecretHelper(client)
	lps.k8sGetterOrCreator = kubernetes.NewGetterOrCreator(client)
	return &lps
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

// Synchronize LogPipeline with ConfigMap of fluentBitDaemonSetHelper sections (Input, Filter and Output).
func (s *syncer) syncSectionsConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := logf.FromContext(ctx)
	cm, err := s.k8sGetterOrCreator.ConfigMap(ctx, s.fluentBitK8sResources.SectionsConfigMap)
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
		fluentBitConfig, err := builder.BuildFluentBitConfig(logPipeline, s.pipelineConfig)
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
	cm, err := s.k8sGetterOrCreator.ConfigMap(ctx, s.fluentBitK8sResources.FilesConfigMap)
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
	oldSecret, err := s.k8sGetterOrCreator.Secret(ctx, s.fluentBitK8sResources.EnvSecret)
	if err != nil {
		return false, err
	}

	newSecret := oldSecret
	newSecret.Data = make(map[string][]byte)

	for _, l := range logPipelines.Items {
		if l.DeletionTimestamp != nil {
			continue
		}
		for _, varRef := range l.Spec.Variables {
			if varRef.ValueFrom.IsSecretRef() {
				err := s.secretHelper.CopySecretData(ctx, varRef.ValueFrom, varRef.Name, newSecret.Data)
				if err != nil {
					log.Error(err, "unable to find secret for environment variable")
					return false, err
				}
			}
		}
		output := l.Spec.Output
		if !output.IsHTTPDefined() {
			continue
		}

		httpOutput := output.HTTP
		if httpOutput.Host.ValueFrom.IsSecretRef() {
			err := s.secretHelper.CopySecretData(ctx, httpOutput.Host.ValueFrom, envvar.GenerateName(l.Name, httpOutput.Host.ValueFrom.SecretKey), newSecret.Data)
			if err != nil {
				log.Error(err, "unable to find secret for http host")
				return false, err
			}
		}
		if httpOutput.User.ValueFrom.IsSecretRef() {
			err := s.secretHelper.CopySecretData(ctx, httpOutput.User.ValueFrom, envvar.GenerateName(l.Name, httpOutput.User.ValueFrom.SecretKey), newSecret.Data)
			if err != nil {
				log.Error(err, "unable to find secret for http user")
				return false, err
			}
		}
		if httpOutput.Password.ValueFrom.IsSecretRef() {
			err := s.secretHelper.CopySecretData(ctx, httpOutput.Password.ValueFrom, envvar.GenerateName(l.Name, httpOutput.Password.ValueFrom.SecretKey), newSecret.Data)
			if err != nil {
				log.Error(err, "unable to find secret for http password")
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

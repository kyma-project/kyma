package sync

import (
	"bytes"
	"context"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	parsersConfigMapKey        = "parsers.conf"
	sectionsConfigMapFinalizer = "FLUENT_BIT_SECTIONS_CONFIG_MAP"
	parserConfigMapFinalizer   = "FLUENT_BIT_PARSERS_CONFIG_MAP"
	secretRefsFinalizer        = "FLUENT_BIT_SECRETS" //nolint:gosec
	filesFinalizer             = "FLUENT_BIT_FILES"
)

type FluentBitDaemonSetConfig struct {
	FluentBitDaemonSetName     types.NamespacedName
	FluentBitSectionsConfigMap types.NamespacedName
	FluentBitParsersConfigMap  types.NamespacedName
	FluentBitFilesConfigMap    types.NamespacedName
	FluentBitEnvSecret         types.NamespacedName
}

type LogPipelineSyncer struct {
	client.Client
	DaemonSetConfig         FluentBitDaemonSetConfig
	EmitterConfig           fluentbit.EmitterConfig
	EnableUnsupportedPlugin bool
}

func NewLogPipelineSyncer(client client.Client,
	daemonSetConfig FluentBitDaemonSetConfig,
	emitterConfig fluentbit.EmitterConfig,
) *LogPipelineSyncer {
	var lps LogPipelineSyncer
	lps.Client = client
	lps.DaemonSetConfig = daemonSetConfig
	lps.EmitterConfig = emitterConfig
	return &lps
}

func (s *LogPipelineSyncer) SyncAll(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := logf.FromContext(ctx)

	sectionsChanged, err := s.syncSectionsConfigMap(ctx, logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync Sections ConfigMap")
		return false, err
	}
	parsersChanged, err := s.syncParsersConfigMap(ctx, logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync Parsers ConfigMap")
		return false, err
	}
	filesChanged, err := s.syncFilesConfigMap(ctx, logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync mounted files")
		return false, err
	}
	secretsChanged, err := s.syncSecretRefs(ctx, logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync secret references")
		return false, err
	}
	err = s.syncEnableUnsupportedPluginMetric(ctx)
	if err != nil {
		return false, err
	}
	return sectionsChanged || parsersChanged || filesChanged || secretsChanged, nil
}

// Synchronize LogPipeline with ConfigMap of FluentBit sections (Input, Filter and Output).
func (s *LogPipelineSyncer) syncSectionsConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := logf.FromContext(ctx)
	cm, err := s.getOrCreateConfigMap(ctx, s.DaemonSetConfig.FluentBitSectionsConfigMap)
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
		fluentBitConfig, err := fluentbit.MergeSectionsConfig(logPipeline, s.EmitterConfig)
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

func (s *LogPipelineSyncer) syncEnableUnsupportedPluginMetric(ctx context.Context) error {
	var logPipelines telemetryv1alpha1.LogPipelineList
	err := s.List(ctx, &logPipelines)
	if err != nil {
		return err
	}

	s.EnableUnsupportedPlugin = updateEnableAllPluginMetrics(&logPipelines)
	return nil
}

// Synchronize LogPipeline with ConfigMap of FluentBit parsers (Parser and MultiLineParser).
func (s *LogPipelineSyncer) syncParsersConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := logf.FromContext(ctx)
	cm, err := s.getOrCreateConfigMap(ctx, s.DaemonSetConfig.FluentBitParsersConfigMap)
	if err != nil {
		return false, err
	}

	changed := false
	var logPipelines telemetryv1alpha1.LogPipelineList

	if logPipeline.DeletionTimestamp != nil {
		if cm.Data != nil && controllerutil.ContainsFinalizer(logPipeline, parserConfigMapFinalizer) {
			log.Info("Deleting fluent bit parsers config")

			err = s.List(ctx, &logPipelines)
			if err != nil {
				return false, err
			}

			fluentBitParsersConfig := fluentbit.MergeParsersConfig(&logPipelines)
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
		err = s.List(ctx, &logPipelines)
		if err != nil {
			return false, err
		}

		fluentBitParsersConfig := fluentbit.MergeParsersConfig(&logPipelines)
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

// Synchronize file references with Fluent Bit files ConfigMap.
func (s *LogPipelineSyncer) syncFilesConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := logf.FromContext(ctx)
	cm, err := s.getOrCreateConfigMap(ctx, s.DaemonSetConfig.FluentBitFilesConfigMap)
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

// Copy referenced secrets to global Fluent Bit environment secret.
func (s *LogPipelineSyncer) syncSecretRefs(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := logf.FromContext(ctx)
	secret, err := s.getOrCreateSecret(ctx, s.DaemonSetConfig.FluentBitEnvSecret)
	if err != nil {
		return false, err
	}

	changed := false
	for _, secretRef := range logPipeline.Spec.SecretRefs {
		var referencedSecret corev1.Secret
		if err = s.Get(ctx, types.NamespacedName{Name: secretRef.Name, Namespace: secretRef.Namespace}, &referencedSecret); err != nil {
			log.Error(err, "Failed reading secret %s from namespace %s", secretRef.Name, secretRef.Namespace)
			continue
		}
		for k, v := range referencedSecret.Data {
			if logPipeline.DeletionTimestamp != nil {
				if _, hasKey := secret.Data[k]; hasKey {
					delete(secret.Data, k)
					controllerutil.RemoveFinalizer(logPipeline, secretRefsFinalizer)
					changed = true
				}
			} else {
				if secret.Data == nil {
					data := make(map[string][]byte)
					data[k] = v
					secret.Data = data
					changed = true
				} else {
					if oldEnv, hasKey := secret.Data[k]; !hasKey || !bytes.Equal(oldEnv, v) {
						secret.Data[k] = v
						changed = true
					}
				}
				if !controllerutil.ContainsFinalizer(logPipeline, secretRefsFinalizer) {
					controllerutil.AddFinalizer(logPipeline, secretRefsFinalizer)
					changed = true
				}
			}
		}
	}

	if !changed {
		return false, nil
	}
	if err = s.Update(ctx, &secret); err != nil {
		return false, err
	}

	return changed, nil
}

func (s *LogPipelineSyncer) getOrCreateConfigMap(ctx context.Context, name types.NamespacedName) (corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name.Name, Namespace: name.Namespace}}
	err := s.getOrCreate(ctx, &cm)
	if err != nil {
		return corev1.ConfigMap{}, err
	}
	return cm, nil
}

func (s *LogPipelineSyncer) getOrCreateSecret(ctx context.Context, name types.NamespacedName) (corev1.Secret, error) {
	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name.Name, Namespace: name.Namespace}}
	err := s.getOrCreate(ctx, &secret)
	if err != nil {
		return corev1.Secret{}, err
	}
	return secret, nil
}

// Gets or creates the given obj in the Kubernetes cluster. obj must be a struct pointer so that obj can be updated with the content returned by the Server.
func (s *LogPipelineSyncer) getOrCreate(ctx context.Context, obj client.Object) error {
	err := s.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	if err != nil && errors.IsNotFound(err) {
		return s.Create(ctx, obj)
	}
	return err
}

func updateEnableAllPluginMetrics(pipelines *telemetryv1alpha1.LogPipelineList) bool {
	enableUnsupportedPlugins := false
	for _, l := range pipelines.Items {
		if l.Spec.EnableUnsupportedPlugins && l.DeletionTimestamp == nil {
			enableUnsupportedPlugins = true
			break
		}
	}
	return enableUnsupportedPlugins
}

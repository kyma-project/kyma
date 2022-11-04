package logpipeline

import (
	"bytes"
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

const (
	sectionsFinalizer = "FLUENT_BIT_SECTIONS_CONFIG_MAP"
	filesFinalizer    = "FLUENT_BIT_FILES"
)

type syncer struct {
	client.Client
	config             Config
	k8sGetterOrCreator *kubernetes.GetterOrCreator
}

func newSyncer(
	client client.Client,
	config Config,
) *syncer {
	var s syncer
	s.Client = client
	s.config = config
	s.k8sGetterOrCreator = kubernetes.NewGetterOrCreator(client)
	return &s
}

func (s *syncer) syncAll(ctx context.Context, newPipeline *telemetryv1alpha1.LogPipeline, allPipelines *telemetryv1alpha1.LogPipelineList) (bool, error) {
	log := logf.FromContext(ctx)

	sectionsChanged, err := s.syncSectionsConfigMap(ctx, newPipeline)
	if err != nil {
		log.Error(err, "Failed to sync sections")
		return false, err
	}

	filesChanged, err := s.syncFilesConfigMap(ctx, newPipeline)
	if err != nil {
		log.Error(err, "Failed to sync mounted files")
		return false, err
	}

	variablesChanged, err := s.syncReferencedSecrets(ctx, allPipelines)
	if err != nil {
		log.Error(err, "Failed to sync referenced secrets")
		return false, err
	}

	return sectionsChanged || filesChanged || variablesChanged, nil
}

func (s *syncer) syncSectionsConfigMap(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	cm, err := s.k8sGetterOrCreator.ConfigMap(ctx, s.config.SectionsConfigMap)
	if err != nil {
		return false, fmt.Errorf("unable to get section configmap: %w", err)
	}

	changed := false
	cmKey := pipeline.Name + ".conf"
	if pipeline.DeletionTimestamp != nil {
		if cm.Data != nil && controllerutil.ContainsFinalizer(pipeline, sectionsFinalizer) {
			delete(cm.Data, cmKey)
			controllerutil.RemoveFinalizer(pipeline, sectionsFinalizer)
			changed = true
		}
	} else {
		newConfig, err := builder.BuildFluentBitConfig(pipeline, s.config.PipelineDefaults)
		if err != nil {
			return false, fmt.Errorf("unable to build section: %w", err)
		}
		if cm.Data == nil {
			cm.Data = map[string]string{cmKey: newConfig}
			changed = true
		} else if oldConfig, hasKey := cm.Data[cmKey]; !hasKey || oldConfig != newConfig {
			cm.Data[cmKey] = newConfig
			changed = true
		}
		if !controllerutil.ContainsFinalizer(pipeline, sectionsFinalizer) {
			controllerutil.AddFinalizer(pipeline, sectionsFinalizer)
			changed = true
		}
	}

	if !changed {
		return false, nil
	}
	if err = s.Update(ctx, &cm); err != nil {
		return false, fmt.Errorf("unable to update section configmap: %w", err)
	}

	return changed, nil
}

func (s *syncer) syncFilesConfigMap(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	cm, err := s.k8sGetterOrCreator.ConfigMap(ctx, s.config.FilesConfigMap)
	if err != nil {
		return false, fmt.Errorf("unable to get files configmap: %w", err)
	}

	changed := false
	for _, file := range pipeline.Spec.Files {
		if pipeline.DeletionTimestamp != nil {
			if _, hasKey := cm.Data[file.Name]; hasKey {
				delete(cm.Data, file.Name)
				controllerutil.RemoveFinalizer(pipeline, filesFinalizer)
				changed = true
			}
		} else {
			if cm.Data == nil {
				cm.Data = map[string]string{file.Name: file.Content}
				changed = true
			} else if oldContent, hasKey := cm.Data[file.Name]; !hasKey || oldContent != file.Content {
				cm.Data[file.Name] = file.Content
				changed = true
			}
			if !controllerutil.ContainsFinalizer(pipeline, filesFinalizer) {
				controllerutil.AddFinalizer(pipeline, filesFinalizer)
				changed = true
			}
		}
	}

	if !changed {
		return false, nil
	}
	if err = s.Update(ctx, &cm); err != nil {
		return false, fmt.Errorf("unable to update files configmap: %w", err)
	}

	return changed, nil
}

func (s *syncer) syncReferencedSecrets(ctx context.Context, logPipelines *telemetryv1alpha1.LogPipelineList) (bool, error) {
	oldSecret, err := s.k8sGetterOrCreator.Secret(ctx, s.config.EnvSecret)
	if err != nil {
		return false, fmt.Errorf("unable to get env secret: %w", err)
	}

	newSecret := oldSecret
	newSecret.Data = make(map[string][]byte)

	for i := range logPipelines.Items {
		if !logPipelines.Items[i].DeletionTimestamp.IsZero() {
			continue
		}

		for _, field := range lookupSecretRefFields(&logPipelines.Items[i]) {
			if copyErr := s.copySecretData(ctx, field.secretKeyRef, field.targetSecretKey, newSecret.Data); copyErr != nil {
				return false, fmt.Errorf("unable to copy secret data: %w", copyErr)
			}
		}
	}

	secretChanged := secretDataEqual(oldSecret.Data, newSecret.Data)
	if !secretChanged {
		return false, nil
	}

	if err = s.Update(ctx, &newSecret); err != nil {
		return false, fmt.Errorf("unable to update env secret: %w", err)
	}
	return secretChanged, nil
}

func (s *syncer) copySecretData(ctx context.Context, sourceRef telemetryv1alpha1.SecretKeyRef, targetKey string, target map[string][]byte) error {
	var source corev1.Secret
	if err := s.Get(ctx, sourceRef.NamespacedName(), &source); err != nil {
		return fmt.Errorf("unable to read secret '%s' from namespace '%s': %w", sourceRef.Name, sourceRef.Namespace, err)
	}

	if val, found := source.Data[sourceRef.Key]; found {
		target[targetKey] = val
		return nil
	}

	return fmt.Errorf("unable to find key '%s' in secret '%s' from namespace '%s'",
		sourceRef.Key,
		sourceRef.Name,
		sourceRef.Namespace)
}

func secretDataEqual(oldSecret, newSecret map[string][]byte) bool {
	if len(newSecret) != len(oldSecret) {
		return true
	}
	for k, newSecretVal := range newSecret {
		if oldSecretVal, ok := oldSecret[k]; !ok || !bytes.Equal(newSecretVal, oldSecretVal) {
			return true
		}
	}
	return false
}

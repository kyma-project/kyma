package fluentbit

import (
	"context"
	"fmt"
	"time"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fs"
	v1 "k8s.io/api/core/v1"

	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type DaemonSetUtils struct {
	client             client.Client
	FluentBitDaemonSet types.NamespacedName
}

func NewDaemonSetUtils(client client.Client, fluenbitDaemonSet types.NamespacedName) *DaemonSetUtils {
	return &DaemonSetUtils{
		client:             client,
		FluentBitDaemonSet: fluenbitDaemonSet,
	}
}

// Delete all Fluent Bit pods to apply new configuration.
func (f *DaemonSetUtils) RestartFluentBit(ctx context.Context, restartsTotal prometheus.Counter) error {
	log := logf.FromContext(ctx)
	log.Info("got counter", "restarts", restartsTotal)
	var ds appsv1.DaemonSet
	if err := f.client.Get(ctx, f.FluentBitDaemonSet, &ds); err != nil {
		log.Error(err, "Failed getting fluent bit DaemonSet")
		return err
	}

	patchedDS := *ds.DeepCopy()
	if patchedDS.Spec.Template.ObjectMeta.Annotations == nil {
		patchedDS.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	patchedDS.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	if err := f.client.Patch(ctx, &patchedDS, client.MergeFrom(&ds)); err != nil {
		log.Error(err, "Failed to patch Fluent Bit to trigger rolling update")
		return err
	}
	restartsTotal.Inc()
	return nil
}

func (f *DaemonSetUtils) IsFluentBitDaemonSetReady(ctx context.Context) (bool, error) {
	log := logf.FromContext(ctx)
	var ds appsv1.DaemonSet
	if err := f.client.Get(ctx, f.FluentBitDaemonSet, &ds); err != nil {
		log.Error(err, "Failed getting fluent bit daemon set")
		return false, err
	}

	generation := ds.Generation
	observedGeneration := ds.Status.ObservedGeneration
	updated := ds.Status.UpdatedNumberScheduled
	desired := ds.Status.DesiredNumberScheduled
	ready := ds.Status.NumberReady

	log.V(1).Info(fmt.Sprintf("Checking fluent bit: updated: %d, desired: %d, ready: %d, generation: %d, observed generation: %d",
		updated, desired, ready, generation, observedGeneration))

	return observedGeneration == generation && updated == desired && ready >= desired, nil
}

func (f *DaemonSetUtils) GetFluentBitConfig(ctx context.Context,
	currentBaseDirectory, fluentBitParsersConfigMapKey string,
	fluentBitConfigMap types.NamespacedName,
	pipelineConfig PipelineConfig,
	pipeline *telemetryv1alpha1.LogPipeline,
	parser *telemetryv1alpha1.LogParser) ([]fs.File, error) {
	var configFiles []fs.File
	fluentBitSectionsConfigDirectory := currentBaseDirectory + "/dynamic"
	fluentBitParsersConfigDirectory := currentBaseDirectory + "/dynamic-parsers"
	fluentBitFilesDirectory := currentBaseDirectory + "/files"

	var generalCm v1.ConfigMap
	if err := f.client.Get(ctx, fluentBitConfigMap, &generalCm); err != nil {
		return []fs.File{}, err
	}
	for key, data := range generalCm.Data {
		configFiles = append(configFiles, fs.File{
			Path: currentBaseDirectory,
			Name: key,
			Data: data,
		})
	}
	var logPipelines telemetryv1alpha1.LogPipelineList
	err := f.client.List(ctx, &logPipelines)
	if err != nil {
		return []fs.File{}, err
	}
	if pipeline != nil {
		logPipelines.Items = append(logPipelines.Items, *pipeline)
	}
	// Build the config from all the exiting pipelines
	for _, logPipeline := range logPipelines.Items {
		configFiles, err = appendConfigFile(configFiles, logPipeline, pipelineConfig, fluentBitSectionsConfigDirectory, fluentBitFilesDirectory)
		if err != nil {
			return []fs.File{}, err
		}
	}

	var parsersConfig string
	var logParsers telemetryv1alpha1.LogParserList
	if err := f.client.List(ctx, &logParsers); err != nil {
		return []fs.File{}, err
	}
	if parser != nil {
		logParsers.Items = append(logParsers.Items, *parser)
	}

	parsersConfig = MergeParsersConfig(&logParsers)
	configFiles = append(configFiles, fs.File{
		Path: fluentBitParsersConfigDirectory,
		Name: fluentBitParsersConfigMapKey,
		Data: parsersConfig,
	})

	return configFiles, nil
}
func appendConfigFile(
	configFiles []fs.File,
	logPipeline telemetryv1alpha1.LogPipeline,
	pipelineConfig PipelineConfig,
	fluentBitSectionsConfigDirectory string,
	fluentBitFilesDirectory string) ([]fs.File, error) {
	for _, file := range logPipeline.Spec.Files {
		configFiles = append(configFiles, fs.File{
			Path: fluentBitFilesDirectory,
			Name: file.Name,
			Data: file.Content,
		})
	}

	sectionsConfig, err := MergeSectionsConfig(&logPipeline, pipelineConfig)
	if err != nil {
		return []fs.File{}, err
	}
	configFiles = append(configFiles, fs.File{
		Path: fluentBitSectionsConfigDirectory,
		Name: logPipeline.Name + ".conf",
		Data: sectionsConfig,
	})
	return configFiles, nil
}

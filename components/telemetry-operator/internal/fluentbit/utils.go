package fluentbit

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/controller/logpipeline/fluentbitconfig"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/utils"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
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
	restartsTotal      prometheus.Counter
}

func NewDaemonSetUtils(client client.Client, fluenbitDaemonSet types.NamespacedName, restartsTotal prometheus.Counter) *DaemonSetUtils {
	return &DaemonSetUtils{
		client:             client,
		FluentBitDaemonSet: fluenbitDaemonSet,
		restartsTotal:      restartsTotal,
	}
}

// RestartFluentBit deletes all Fluent Bit pods to apply new configuration
func (f *DaemonSetUtils) RestartFluentBit(ctx context.Context) error {
	log := logf.FromContext(ctx)
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
	f.restartsTotal.Inc()
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
	pipelineConfig fluentbitconfig.PipelineConfig,
	pipeline *telemetryv1alpha1.LogPipeline,
	parser *telemetryv1alpha1.LogParser) ([]utils.File, error) {
	var configFiles []utils.File
	fluentBitSectionsConfigDirectory := currentBaseDirectory + "/dynamic"
	fluentBitParsersConfigDirectory := currentBaseDirectory + "/dynamic-parsers"
	fluentBitFilesDirectory := currentBaseDirectory + "/files"

	var generalCm v1.ConfigMap
	var logParsers telemetryv1alpha1.LogParserList
	var err error

	if err := f.client.Get(ctx, fluentBitConfigMap, &generalCm); err != nil {
		return []utils.File{}, err
	}
	for key, data := range generalCm.Data {
		configFiles = append(configFiles, utils.File{
			Path: currentBaseDirectory,
			Name: key,
			Data: data,
		})
	}
	// If validating pipeline then check pipelines + parsers
	if pipeline != nil {
		configFiles, err = appendFluentBitConfigFile(configFiles, *pipeline, pipelineConfig, fluentBitSectionsConfigDirectory, fluentBitFilesDirectory)
		if err != nil {
			return []utils.File{}, err
		}
		if err = f.client.List(ctx, &logParsers); err != nil {
			return []utils.File{}, err
		}
		parsersConfig := fluentbitconfig.MergeParsersConfig(&logParsers)
		configFiles = append(configFiles, utils.File{
			Path: fluentBitParsersConfigDirectory,
			Name: fluentBitParsersConfigMapKey,
			Data: parsersConfig,
		})

		return configFiles, nil
	}

	if parser != nil {
		logParsers.Items = appendUniqueParsers(logParsers.Items, parser)
		parsersConfig := fluentbitconfig.MergeParsersConfig(&logParsers)
		configFiles = append(configFiles, utils.File{
			Path: fluentBitParsersConfigDirectory,
			Name: fluentBitParsersConfigMapKey,
			Data: parsersConfig,
		})

		return configFiles, nil
	}

	return []utils.File{}, fmt.Errorf("either Pipeline or Parser should be passed to be validated")
}

func appendUniqueParsers(logParsers []telemetryv1alpha1.LogParser, parser *telemetryv1alpha1.LogParser) []telemetryv1alpha1.LogParser {
	for _, l := range logParsers {
		if l.Name == parser.Name {
			l = *parser
			return logParsers
		}
	}
	return append(logParsers, *parser)
}

func appendFluentBitConfigFile(
	configFiles []utils.File,
	logPipeline telemetryv1alpha1.LogPipeline,
	pipelineConfig fluentbitconfig.PipelineConfig,
	fluentBitSectionsConfigDirectory string,
	fluentBitFilesDirectory string) ([]utils.File, error) {
	for _, file := range logPipeline.Spec.Files {
		configFiles = append(configFiles, utils.File{
			Path: fluentBitFilesDirectory,
			Name: file.Name,
			Data: file.Content,
		})
	}

	sectionsConfig, err := fluentbitconfig.MergeSectionsConfig(&logPipeline, pipelineConfig)
	if err != nil {
		return []utils.File{}, err
	}

	configFiles = append(configFiles, utils.File{
		Path: fluentBitSectionsConfigDirectory,
		Name: logPipeline.Name + ".conf",
		Data: sectionsConfig,
	})
	return configFiles, nil
}

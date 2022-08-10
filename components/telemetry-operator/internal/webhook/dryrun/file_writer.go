package dryrun

import (
	"context"
	"path/filepath"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type fileWriter struct {
	client client.Client
	config *Config
}

type fileCleaner func()

func (f *fileWriter) preparePipelineDryRun(ctx context.Context, workDir string, pipeline *telemetryv1alpha1.LogPipeline) (fileCleaner, error) {
	if err := MakeDir(workDir); err != nil {
		return nil, err
	}
	if err := f.writeConfig(ctx, workDir); err != nil {
		return nil, err
	}
	if err := f.writeFiles(pipeline, workDir); err != nil {
		return nil, err
	}
	if err := f.writeSections(pipeline, workDir); err != nil {
		return nil, err
	}
	if err := f.writeParsers(ctx, workDir); err != nil {
		return nil, err
	}

	cleaner := func() {
		if err := RemoveAll(workDir); err != nil {
			log := logf.FromContext(ctx)
			log.Error(err, "Failed to remove Fluent Bit config directory")
		}
	}
	return cleaner, nil
}

func (f *fileWriter) writeConfig(ctx context.Context, basePath string) error {
	var cm v1.ConfigMap
	var err error
	if err = f.client.Get(ctx, f.config.FluentBitConfigMapName, &cm); err != nil {
		return err
	}
	for key, data := range cm.Data {
		err = WriteFile(filepath.Join(basePath, key), data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *fileWriter) writeFiles(pipeline *telemetryv1alpha1.LogPipeline, basePath string) error {
	filesDir := filepath.Join(basePath, "files")
	if err := MakeDir(filesDir); err != nil {
		return err
	}

	for _, file := range pipeline.Spec.Files {
		err := WriteFile(filepath.Join(filesDir, file.Name), file.Content)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *fileWriter) writeSections(pipeline *telemetryv1alpha1.LogPipeline, basePath string) error {
	dynamicDir := filepath.Join(basePath, "dynamic")
	if err := MakeDir(dynamicDir); err != nil {
		return err
	}

	sectionsConfig, err := fluentbit.MergeSectionsConfig(pipeline, f.config.PipelineConfig)
	if err != nil {
		return err
	}
	return WriteFile(filepath.Join(dynamicDir, pipeline.Name+".conf"), sectionsConfig)
}

func (f *fileWriter) writeParsers(ctx context.Context, basePath string) error {
	dynamicParsersDir := filepath.Join(basePath, "dynamic-parsers")
	if err := MakeDir(dynamicParsersDir); err != nil {
		return err
	}

	var logParsers telemetryv1alpha1.LogParserList
	if err := f.client.List(ctx, &logParsers); err != nil {
		return err
	}
	parsersConfig := fluentbit.MergeParsersConfig(&logParsers)

	return WriteFile(filepath.Join(dynamicParsersDir, "parsers.conf"), parsersConfig)
}

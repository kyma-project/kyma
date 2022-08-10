package dryrun

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const errDescription = "Validation of the supplied configuration failed with the following reason: "

func fluentBitArgs() []string {
	return []string{"--dry-run", "--quiet"}
}

type Config struct {
	FluentBitBinPath       string
	FluentBitPluginDir     string
	FluentBitConfigMapName types.NamespacedName
	PipelineConfig         fluentbit.PipelineConfig
}

type DryRunner struct {
	client client.Client
	config *Config
}

func NewDryRunner(c client.Client, config *Config) *DryRunner {
	return &DryRunner{
		client: c,
		config: config,
	}
}

func (d *DryRunner) DryRunParser(ctx context.Context, parser *telemetryv1alpha1.LogParser) error {
	// TODO Parsers
	//workDir := "/tmp/dry-run" + uuid.New().String()
	//path := filepath.Join(workDir, "dynamic-parsers", "parsers.conf")
	//args := append(fluentBitArgs(), "--parser", path)
	//return d.run(ctx, args)
	return nil
}

// Will run:  fluent-bit/bin/fluent-bit --dry-run --quiet --config /tmp/dry-run018231-123123-123123-123/fluent-bit.conf -e plugin1 -e plugin2 -e plugin3
func (d *DryRunner) DryRunPipeline(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) error {
	workDir, err := createNewWorkDir()
	if err != nil {
		return err
	}
	if err = d.writeConfig(ctx, workDir); err != nil {
		return err
	}
	if err = d.writeFiles(pipeline, workDir); err != nil {
		return err
	}
	if err = d.writeSections(pipeline, workDir); err != nil {
		return err
	}
	if err = d.writeParsers(ctx, workDir); err != nil {
		return err
	}
	defer func() {
		if err = RemoveAll(workDir); err != nil {
			log := logf.FromContext(ctx)
			log.Error(err, "Failed to remove Fluent Bit config directory")
		}
	}()

	path := filepath.Join(workDir, "fluent-bit.conf")
	args := append(fluentBitArgs(), "--config", path)

	return d.runCmd(ctx, args)
}

func createNewWorkDir() (string, error) {
	var newWorkDir = "/tmp/dry-run-" + uuid.New().String()
	err := MakeDir(newWorkDir)
	if err != nil {
		return "", err
	}
	return newWorkDir, nil
}

func (d *DryRunner) writeConfig(ctx context.Context, basePath string) error {
	var cm v1.ConfigMap
	var err error
	if err = d.client.Get(ctx, d.config.FluentBitConfigMapName, &cm); err != nil {
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

func (d *DryRunner) writeFiles(pipeline *telemetryv1alpha1.LogPipeline, basePath string) error {
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

func (d *DryRunner) writeSections(pipeline *telemetryv1alpha1.LogPipeline, basePath string) error {
	dynamicDir := filepath.Join(basePath, "dynamic")
	if err := MakeDir(dynamicDir); err != nil {
		return err
	}

	sectionsConfig, err := fluentbit.MergeSectionsConfig(pipeline, d.config.PipelineConfig)
	if err != nil {
		return err
	}
	return WriteFile(filepath.Join(dynamicDir, pipeline.Name+".conf"), sectionsConfig)
}

func (d *DryRunner) writeParsers(ctx context.Context, basePath string) error {
	dynamicParsersDir := filepath.Join(basePath, "dynamic-parsers")
	if err := MakeDir(dynamicParsersDir); err != nil {
		return err
	}

	var logParsers telemetryv1alpha1.LogParserList
	if err := d.client.List(ctx, &logParsers); err != nil {
		return err
	}
	parsersConfig := fluentbit.MergeParsersConfig(&logParsers)

	return WriteFile(filepath.Join(dynamicParsersDir, "parsers.conf"), parsersConfig)
}

func (d *DryRunner) runCmd(ctx context.Context, args []string) error {
	var plugins []string
	files, err := ioutil.ReadDir(d.config.FluentBitPluginDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		plugins = append(plugins, filepath.Join(d.config.FluentBitPluginDir, f.Name()))
	}
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		args = append(args, "-e", plugin)
	}

	cmd := exec.CommandContext(ctx, d.config.FluentBitBinPath, args...)
	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	if err != nil {
		if strings.Contains(out, "error") || strings.Contains(out, "Error") {
			return errors.New(errDescription + extractError(out))
		}
		return fmt.Errorf("error while validating Fluent Bit config: %v", err.Error())
	}
	return nil
}

func extractError(output string) string {
	// Found in https://github.com/acarl005/stripansi/blob/master/stripansi.go#L7
	rColors := regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")
	output = rColors.ReplaceAllString(output, "")

	// Error pattern: [<time>] [error] [config] error in path/to/file.conf:3: <msg>
	r1 := regexp.MustCompile(`.*(?P<label>\[error]\s\[config].+:\s)(?P<description>.+)`)
	if r1Matches := r1.FindStringSubmatch(output); r1Matches != nil {
		return r1Matches[2] // 0: complete output, 1: label, 2: description
	}

	// Error pattern: [<time>] [error] [config] <msg>
	r2 := regexp.MustCompile(`.*(?P<label>\[error]\s\[config]\s)(?P<description>.+)`)
	if r2Matches := r2.FindStringSubmatch(output); r2Matches != nil {
		return r2Matches[2] // 0: complete output, 1: label, 2: description
	}

	// Error pattern: [<time>] [error] [parser] <msg> in path/to/file.conf
	r3 := regexp.MustCompile(`.*(?P<label>\[error]\s\[parser]\s)(?P<description>.+)(\sin.+)`)
	if r3Matches := r3.FindStringSubmatch(output); r3Matches != nil {
		return r3Matches[2] // 0: complete output, 1: label, 2: description 3: file name
	}

	// Error pattern: [<time>] [error] <msg>
	r4 := regexp.MustCompile(`.*(?P<label>\[error]\s)(?P<description>.+)`)
	if r4Matches := r4.FindStringSubmatch(output); r4Matches != nil {
		return r4Matches[2] // 0: complete output, 1: label, 2: description
	}

	// Error pattern: error<msg>
	r5 := regexp.MustCompile(`error.+`)
	return r5.FindString(output)
}

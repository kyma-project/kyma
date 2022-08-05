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
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errDescription = "Validation of the supplied configuration failed with the following reason: "
	// From https://github.com/acarl005/stripansi/blob/master/stripansi.go#L7
	ansiColorsRegex = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

	fluentBitConfigDirectory     = "/tmp/dry-run"
	fluentBitParsersConfigMapKey = "parsers.conf"
)

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

func (d *DryRunner) DryRunPipeline(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) error {
	currentBaseDirectory := fluentBitConfigDirectory + uuid.New().String()
	path := filepath.Join(currentBaseDirectory, "fluent-bit.conf")
	args := append(fluentBitArgs(), "--config", path)
	return d.run(ctx, args)
}

func (d *DryRunner) DryRunParser(ctx context.Context, parser *telemetryv1alpha1.LogParser) error {
	currentBaseDirectory := fluentBitConfigDirectory + uuid.New().String()
	path := filepath.Join(currentBaseDirectory, "dynamic-parsers", "parsers.conf")
	args := append(fluentBitArgs(), "--parser", path)
	return d.run(ctx, args)
}

func (d *DryRunner) prepareFluentBitConfig(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline, parser *telemetryv1alpha1.LogParser) ([]utils.File, error) {
	currentBaseDirectory := fluentBitConfigDirectory + uuid.New().String()

	fluentBitSectionsConfigDirectory := currentBaseDirectory + "/dynamic"
	fluentBitParsersConfigDirectory := currentBaseDirectory + "/dynamic-parsers"
	fluentBitFilesDirectory := currentBaseDirectory + "/files"

	var cm v1.ConfigMap
	var err error
	if err := d.client.Get(ctx, d.config.FluentBitConfigMapName, &cm); err != nil {
		return []utils.File{}, err
	}

	var configFiles []utils.File
	for key, data := range cm.Data {
		configFiles = append(configFiles, utils.File{
			Path: currentBaseDirectory,
			Name: key,
			Data: data,
		})
	}

	var logParsers telemetryv1alpha1.LogParserList
	// If validating pipeline then check pipelines + parsers
	if pipeline != nil {
		configFiles, err = appendFluentBitConfigFile(configFiles, *pipeline, d.config.PipelineConfig, fluentBitSectionsConfigDirectory, fluentBitFilesDirectory)
		if err != nil {
			return []utils.File{}, err
		}
		if err = d.client.List(ctx, &logParsers); err != nil {
			return []utils.File{}, err
		}
		parsersConfig := fluentbit.MergeParsersConfig(&logParsers)
		configFiles = append(configFiles, utils.File{
			Path: fluentBitParsersConfigDirectory,
			Name: fluentBitParsersConfigMapKey,
			Data: parsersConfig,
		})

		return configFiles, nil
	}

	if parser != nil {
		logParsers.Items = appendUniqueParsers(logParsers.Items, parser)
		parsersConfig := fluentbit.MergeParsersConfig(&logParsers)
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
	pipelineConfig fluentbit.PipelineConfig,
	fluentBitSectionsConfigDirectory string,
	fluentBitFilesDirectory string) ([]utils.File, error) {
	for _, file := range logPipeline.Spec.Files {
		configFiles = append(configFiles, utils.File{
			Path: fluentBitFilesDirectory,
			Name: file.Name,
			Data: file.Content,
		})
	}

	sectionsConfig, err := fluentbit.MergeSectionsConfig(&logPipeline, pipelineConfig)
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

func (d *DryRunner) run(ctx context.Context, args []string) error {
	plugins, err := listPlugins(d.config.FluentBitPluginDir)
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		args = append(args, "-e", plugin)
	}

	out, err := runCmd(ctx, d.config.FluentBitBinPath, args...)
	if err != nil {
		if strings.Contains(out, "error") || strings.Contains(out, "Error") {
			return errors.New(errDescription + extractError(out))
		}
		return fmt.Errorf("error while validating Fluent Bit config: %v", err.Error())
	}
	return nil
}

func listPlugins(pluginPath string) ([]string, error) {
	var plugins []string
	files, err := ioutil.ReadDir(pluginPath)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		plugins = append(plugins, filepath.Join(pluginPath, f.Name()))
	}
	return plugins, err
}

func runCmd(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	return out, err
}

// extractError extracts the error message from the output of fluent-bit
// Thereby, it supports the following error patterns:
// 1. [<time>] [error] [config] error in path/to/file.conf:3: <msg>
// 2. [<time>] [error] [config] <msg>
// 3. [<time>] [error] [parser] <msg> in path/to/file.conf
// 4. [<time>] [error] <msg>
// 5. error<msg>
func extractError(output string) string {
	rColors := regexp.MustCompile(ansiColorsRegex)
	output = rColors.ReplaceAllString(output, "")

	r1 := regexp.MustCompile(`.*(?P<label>\[error]\s\[config].+:\s)(?P<description>.+)`)
	if r1Matches := r1.FindStringSubmatch(output); r1Matches != nil {
		return r1Matches[2] // 0: complete output, 1: label, 2: description
	}

	r2 := regexp.MustCompile(`.*(?P<label>\[error]\s\[config]\s)(?P<description>.+)`)
	if r2Matches := r2.FindStringSubmatch(output); r2Matches != nil {
		return r2Matches[2] // 0: complete output, 1: label, 2: description
	}

	r3 := regexp.MustCompile(`.*(?P<label>\[error]\s\[parser]\s)(?P<description>.+)(\sin.+)`)
	if r3Matches := r3.FindStringSubmatch(output); r3Matches != nil {
		return r3Matches[2] // 0: complete output, 1: label, 2: description 3: file name
	}

	r4 := regexp.MustCompile(`.*(?P<label>\[error]\s)(?P<description>.+)`)
	if r4Matches := r4.FindStringSubmatch(output); r4Matches != nil {
		return r4Matches[2] // 0: complete output, 1: label, 2: description
	}

	r5 := regexp.MustCompile(`error.+`)
	return r5.FindString(output)
}

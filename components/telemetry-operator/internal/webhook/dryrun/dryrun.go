package dryrun

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func dryRunArgs() []string {
	return []string{"--dry-run", "--quiet"}
}

type Config struct {
	FluentBitBinPath       string
	FluentBitPluginDir     string
	FluentBitConfigMapName types.NamespacedName
	PipelineConfig         fluentbit.PipelineConfig
}

type DryRunner struct {
	fileWriter    fileWriter
	commandRunner commandRunner
	config        *Config
}

func NewDryRunner(c client.Client, config *Config) *DryRunner {
	return &DryRunner{
		fileWriter:    &fileWriterImpl{client: c, config: config},
		commandRunner: &commandRunnerImpl{},
		config:        config,
	}
}

func (d *DryRunner) DryRunParser(ctx context.Context, parser *telemetryv1alpha1.LogParser) error {
	workDir := newWorkDirPath()
	cleanup, err := d.fileWriter.prepareParserDryRun(ctx, workDir, parser)
	if err != nil {
		return err
	}
	defer cleanup()

	path := filepath.Join(workDir, "dynamic-parsers", "parsers.conf")
	args := append(dryRunArgs(), "--parser", path)
	return d.runCmd(ctx, args)
}

func (d *DryRunner) DryRunPipeline(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) error {
	workDir := newWorkDirPath()
	cleanup, err := d.fileWriter.preparePipelineDryRun(ctx, workDir, pipeline)
	if err != nil {
		return err
	}
	defer cleanup()

	path := filepath.Join(workDir, "fluent-bit.conf")
	args := dryRunArgs()
	args = append(args, "--config", path)
	externalPluginArgs, err := d.externalPluginArgs()
	if err != nil {
		return err
	}
	args = append(args, externalPluginArgs...)

	return d.runCmd(ctx, args)
}

func (d *DryRunner) externalPluginArgs() ([]string, error) {
	if d.config.FluentBitPluginDir == "" {
		return nil, nil
	}

	var plugins []string
	files, err := os.ReadDir(d.config.FluentBitPluginDir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		plugins = append(plugins, filepath.Join(d.config.FluentBitPluginDir, f.Name()))
	}

	var args []string
	for _, plugin := range plugins {
		args = append(args, "-e", plugin)
	}
	return args, nil
}

func (d *DryRunner) runCmd(ctx context.Context, args []string) error {
	outBytes, err := d.commandRunner.run(ctx, d.config.FluentBitBinPath, args...)
	out := string(outBytes)
	if err != nil {
		if strings.Contains(out, "error") || strings.Contains(out, "Error") {
			return fmt.Errorf("failed to validate the supplied configuration: %s", extractError(out))
		}
		return fmt.Errorf("failed to execute dryrun: %v", err.Error())
	}

	return nil
}

func newWorkDirPath() string {
	return "/tmp/dry-run-" + uuid.New().String()
}

func extractError(output string) string {
	// Found in https://github.com/acarl005/stripansi/blob/master/stripansi.go#L7
	rColors := regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\a|(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~])")
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

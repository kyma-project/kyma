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
)

const (
	errDescription = "Validation of the supplied configuration failed with the following reason: "
	// From https://github.com/acarl005/stripansi/blob/master/stripansi.go#L7
	ansiColorsRegex = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
)

func fluentBitArgs() []string {
	return []string{"--dry-run", "--quiet"}
}

type DryRunner struct {
	fluentBitPath   string
	pluginDirectory string
}

func NewDryRunner(fluentBitPath string, pluginDirectory string) DryRunner {
	return DryRunner{
		fluentBitPath:   fluentBitPath,
		pluginDirectory: pluginDirectory,
	}
}

func (v DryRunner) RunConfig(ctx context.Context, configsDirectory string) error {
	path := filepath.Join(configsDirectory, "fluent-bit.conf")
	args := append(fluentBitArgs(), "--config", path)
	return v.run(ctx, args)
}

func (v DryRunner) RunParser(ctx context.Context, configsDirectory string) error {
	path := filepath.Join(configsDirectory, "dynamic-parsers", "parsers.conf")
	args := append(fluentBitArgs(), "--parser", path)
	return v.run(ctx, args)
}

func (v *DryRunner) run(ctx context.Context, args []string) error {
	plugins, err := listPlugins(v.pluginDirectory)
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		args = append(args, "-e", plugin)
	}

	out, err := runCmd(ctx, v.fluentBitPath, args...)
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

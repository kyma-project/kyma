package validation

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

//go:generate mockery --name ConfigValidator --filename config_validator.go
type ConfigValidator interface {
	RunCmd(ctx context.Context, name string, args ...string) (string, error)
	Validate(ctx context.Context, configFilePath string) error
}

type configValidator struct {
	FluentBitPath   string
	PluginDirectory string
}

func NewConfigValidator(fluentBitPath string, pluginDirectory string) ConfigValidator {
	return &configValidator{
		FluentBitPath:   fluentBitPath,
		PluginDirectory: pluginDirectory,
	}
}

func (v *configValidator) RunCmd(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	return out, err
}

func (v *configValidator) Validate(ctx context.Context, configFilePath string) error {

	fluentBitArgs := []string{"--dry-run", "--quiet", "--config", configFilePath}

	if strings.Contains(configFilePath, "parsers.conf") {
		fluentBitArgs = []string{"--dry-run", "--quiet", "-R", configFilePath}
	}
	plugins, err := listPlugins(v.PluginDirectory)
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		fluentBitArgs = append(fluentBitArgs, "-e", plugin)
	}

	out, err := v.RunCmd(ctx, v.FluentBitPath, fluentBitArgs...)
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

package fluentbit

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

const (
	fluentBitBinPath = "fluent-bit/bin/fluent-bit"
	errDescription   = "The Fluent Bit dry-run could not validate the provided Log Pipeline: "
	// From https://github.com/acarl005/stripansi/blob/master/stripansi.go#L7
	ansiColorsRegex = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
)

//go:generate mockery --name ConfigValidator --filename config_validator.go
type ConfigValidator interface {
	RunCmd(ctx context.Context, name string, args ...string) (string, error)
	Validate(ctx context.Context, configFilePath string) error
}

type configValidator struct{}

func NewConfigValidator() ConfigValidator {
	return &configValidator{}
}

func (v *configValidator) RunCmd(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	return out, err
}

func (v *configValidator) Validate(ctx context.Context, configFilePath string) error {
	out, err := v.RunCmd(ctx, fluentBitBinPath, "--dry-run", "--quiet", "--config", configFilePath)
	if err != nil {
		if strings.Contains(out, "Error") {
			return errors.New(errDescription + extractError(out))
		}
		return errors.New(fmt.Sprintf("Error while validating Fluent Bit config: %v", err))
	}

	return nil
}

// extractError extracts the error message from the output of fluent-bit
// Thereby, it supports the following error patterns:
// 1. Error <msg>\nError: Configuration file contains errors. Aborting
// 2. Error: <msg>. Aborting
// 3. [<time>] [  Error] File <filename>\n[<time>] [  Error] Error in line 4: <msg> Error: Configuration file contains errors. Aborting
func extractError(output string) string {

	rColors := regexp.MustCompile(ansiColorsRegex)
	output = rColors.ReplaceAllString(output, "")

	r1 := regexp.MustCompile(`(?P<description>Error[^\]].+)\n(?P<label>Error:.+)`)
	if r1Matches := r1.FindStringSubmatch(output); r1Matches != nil {
		return r1Matches[1] // 0: complete output, 1: description, 2: label
	}

	r2 := regexp.MustCompile(`.*(?P<label>Error:\s)(?P<description>.+\.)`)
	if r2Matches := r2.FindStringSubmatch(output); r2Matches != nil {
		return r2Matches[2] // 0: complete output, 1: label, 2: description
	}

	r3 := regexp.MustCompile(`Error\s.+`)
	return r3.FindString(output)
}

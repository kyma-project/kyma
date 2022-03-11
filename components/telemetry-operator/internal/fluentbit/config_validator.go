package fluentbit

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const (
	fluentBitBinPath = "/bin/fluent-bit"
)

type cmdRunner struct{}

//func NewCmdRunner() CmdRunner {
//	return &cmdRunner{}
//}

type CmdRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

func (e *cmdRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	outBytes, err := cmd.CombinedOutput()
	out := string(outBytes)
	return out, err
}

func Validate(ctx context.Context, configFilePath string) error {

	cmd := cmdRunner{}
	res, err := cmd.Run(ctx, fluentBitBinPath, "-D", "-q", "-c", configFilePath)
	if err != nil {
		if strings.Contains(res, "Error") {
			return errors.New(res)
		}
		return errors.New(fmt.Sprintf("Error while validating Fluent Bit config: %v", err))
	}

	return nil
}

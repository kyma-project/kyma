package fluentbit

import (
	"context"
	"errors"
	"os/exec"
	"strings"
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

func Validate(path string) error {
	binPath := "/bin/fluent-bit"
	flags := []string{"-D", "-q", "-c", path}

	cmd := cmdRunner{}
	res, err := cmd.Run(context.TODO(), binPath, flags...)
	if err != nil {
		return err
	}
	if strings.Contains(res, "Error") {
		err := errors.New(res)
		return err
	}

	return nil
}

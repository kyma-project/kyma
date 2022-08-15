package dryrun

import (
	"context"
	"os/exec"
)

//go:generate mockery --name commandRunner --filename command_runner.go
type commandRunner interface {
	run(ctx context.Context, command string, args ...string) ([]byte, error)
}

type commandRunnerImpl struct{}

func (r *commandRunnerImpl) run(ctx context.Context, command string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, command, args...).CombinedOutput()
}

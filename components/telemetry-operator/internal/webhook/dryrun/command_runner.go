package dryrun

import (
	"context"
	"os/exec"
)

//go:generate mockery --name commandRunner --filename mocks.go
type commandRunner interface {
	run(ctx context.Context, command string, args ...string) ([]byte, error)
}

type realCommandRunner struct{}

func (r *realCommandRunner) run(ctx context.Context, command string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, command, args...).CombinedOutput()
}

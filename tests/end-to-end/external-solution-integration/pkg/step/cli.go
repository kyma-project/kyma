package step

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// Execute behavior is based on chose cleanup method. It is intended to be used with AddFlags
func (r *Runner) Execute(steps []Step) error {
	r.log.Infof("Cleanup mode: %s", r.cleanup)
	var err error
	switch r.cleanup {
	case CleanupModeNo:
		err = r.Run(steps, true)
	case CleanupModeOnly:
		r.Cleanup(steps)
	case CleanupModeYes:
		err = r.Run(steps, false)
	}
	return err
}

// AddFlags add CLI flags so user may control runner behaviour easily
func (r *Runner) AddFlags(set *pflag.FlagSet) {
	set.Var(&r.cleanup, "cleanup", "Cleanup mode. Allowed values: yes/no/only")
}

// CleanupMode says how runner should execute cleanup
type CleanupMode string

const (
	// CleanupModeNo - Don't execute cleanup
	CleanupModeNo CleanupMode = "no"
	// CleanupModeOnly - Don't run steps, only cleanup
	CleanupModeOnly CleanupMode = "only"
	// Execute both steps and cleanup
	CleanupModeYes CleanupMode = "yes"
)

// String implements pflag.Value.String
func (m CleanupMode) String() string {
	return string(m)
}

// Set implements pflag.Value.Set
func (m *CleanupMode) Set(v string) error {
	switch CleanupMode(v) {
	case CleanupModeNo, CleanupModeYes, CleanupModeOnly:
	default:
		return errors.Errorf("invalid cleanup value: %s", v)
	}
	*m = CleanupMode(v)
	return nil
}

// Type implements pflag.Value.Type
func (m CleanupMode) Type() string {
	return "runner mode"
}

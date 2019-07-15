package step

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// Execute behavior is based on chose cleanup method. It is intended to be used with AddFlags
func (r *Runner) Execute(steps []Step) error {
	var err error
	switch r.cleanup {
	case CleanupMode_No:
		err = r.Run(steps, true)
	case CleanupMode_Yes:
		err = r.Run(steps, false)
	case CleanupMode_Only:
		r.Cleanup(steps)
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
	// Don't execute cleanup
	CleanupMode_No   CleanupMode = "no"
	// Don't run steps, only cleanup
	CleanupMode_Only CleanupMode = "only"
	// Execute both steps and cleanup
	CleanupMode_Yes  CleanupMode = "yes"
)

// String implements pflag.Value.String
func (m CleanupMode) String() string {
	return string(m)
}

// Set implements pflag.Value.Set
func (m *CleanupMode) Set(v string) error {
	var result CleanupMode
	switch CleanupMode(v) {
	case CleanupMode_No:
		result = CleanupMode_No
	case CleanupMode_Yes:
		result = CleanupMode_Yes
	case CleanupMode_Only:
		result = CleanupMode_Only
	default:
		return errors.Errorf("invalid cleanup value: %s", v)
	}
	m = &result
	return nil
}

// Type implements pflag.Value.Type
func (m CleanupMode) Type() string {
	return "runner mode"
}

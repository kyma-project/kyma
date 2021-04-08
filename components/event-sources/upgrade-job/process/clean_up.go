package process

import (
	"github.com/pkg/errors"
)

var _ Step = &CleanUpStep{}

type CleanUpStep struct {
	name    string
	process *Process
}

func NewCleanUp(p *Process) CleanUpStep {
	return CleanUpStep{
		name:    "Cleanup of the backed up Config Map",
		process: p,
	}
}

func (s CleanUpStep) Do() error {
	err := s.process.Clients.ConfigMap.Delete(BackedUpConfigMapNamespace, BackedUpConfigMapName)
	if err != nil {
		return errors.Wrapf(err, "failed to delete backed up config map %s/%s", BackedUpConfigMapNamespace, BackedUpConfigMapName)
	}
	s.process.Logger.Infof("Step: %s, cleaned up configmap: %s/%s", s.ToString(), BackedUpConfigMapNamespace, BackedUpConfigMapName)
	return nil
}

func (s CleanUpStep) ToString() string {
	return s.name
}

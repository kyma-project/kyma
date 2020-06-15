package steps

import (
	"errors"
)

type uninstallStep struct {
	step
}

// Run method for uninstallStep triggers step delete via helm. Uninstall should not be retried, hence no error is returned.
func (s uninstallStep) Run() error {

	if deleteErr := s.helmClient.UninstallRelease(s.GetNamespacedName()); deleteErr != nil {
		return errors.New("Helm delete error: " + deleteErr.Error())
	}

	return nil
}

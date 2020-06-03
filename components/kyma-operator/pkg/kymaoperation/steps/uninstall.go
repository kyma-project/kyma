package steps

import "errors"

type uninstallStep struct {
	step
}

// Run method for uninstallStep triggers step delete via helm. Uninstall should not be retried, hence no error is returned.
func (s uninstallStep) Run() error {

	uninstallReleaseResponse, deleteErr := s.helmClient.DeleteRelease(s.component.GetReleaseName())

	if deleteErr != nil {
		return errors.New("Helm delete error: " + deleteErr.Error())
	}

	s.helmClient.PrintRelease(uninstallReleaseResponse.Release)
	return nil
}

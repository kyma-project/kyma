package steps

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type uninstallStep struct {
	step
}

// Run method for uninstallStep triggers step delete via helm. Uninstall should not be retried, hence no error is returned.
func (s uninstallStep) Run() error {

	isPresent, err := s.helmClient.IsReleasePresent(s.GetNamespacedName())

	if !isPresent {
		if err != nil {
			return errors.New("Helm delete error: While checking release status: " + err.Error())
		}
		log.Warnf("Release %s not found: skipping uninstall step", s.GetReleaseName())
		return nil
	}

	if deleteErr := s.helmClient.UninstallRelease(s.GetNamespacedName()); deleteErr != nil {
		if strings.Contains(deleteErr.Error(), driver.ErrReleaseNotFound.Error()) {
			log.Warnf("Release %s not found: skipping uninstall step", s.GetReleaseName())
			return nil
		}
		return errors.New("Helm delete error: " + deleteErr.Error())
	}

	return nil
}

package steps

import (
	"errors"
	"fmt"
	"log"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
)

type upgradeStep struct {
	installStep
	// deployedRevision    int //todo: deployed revision is refreshed before rollback, remove this field.
	rollbackWaitTimeSec uint32
}

// Run method for upgradeStep triggers step upgrade via helm
func (s upgradeStep) Run() error {

	chartDir, err := s.sourceGetter.SrcDirFor(s.component)
	if err != nil {
		return err
	}

	upgradeResp, upgradeErr := s.helmClient.UpgradeRelease(
		chartDir,
		s.component.Namespace,
		s.component.GetReleaseName(),
		s.overrideData.ForRelease(s.component.GetReleaseName()))

	if upgradeErr != nil {
		return s.onError(upgradeErr)
	}

	s.helmClient.PrintRelease(upgradeResp)

	return nil
}

func (s upgradeStep) onError(upgradeErr error) error {

	upgradeErrMsg := fmt.Sprintf("Helm upgrade error: %s", upgradeErr.Error())

	log.Println(fmt.Sprintf("Helm upgrade of release \"%s\" failed. Finding last known deployed revision to rollback to.", s.component.GetReleaseName()))

	rollbackTo, err := s.helmClient.ReleaseDeployedRevision(s.component.Namespace, s.component.GetReleaseName())
	if err != nil {
		deployedRevisionErr := fmt.Sprintf("an error occurred while looking for a deployed %s release: %s", s.component.GetReleaseName(), err.Error())
		return errors.New(fmt.Sprintf("%s \n %s \n", upgradeErrMsg, deployedRevisionErr))
	}

	log.Println(fmt.Sprintf("Performing rollback to last known deployed revision: %d", rollbackTo))

	_, err = s.helmClient.RollbackRelease(s.component.Namespace, s.component.GetReleaseName(), rollbackTo)
	if err != nil {
		rollbackErrMsg := fmt.Sprintf("Helm rollback of release \"%s\" failed with an error: %s", s.component.GetReleaseName(), err.Error())
		return errors.New(fmt.Sprintf("%s \n %s \n", upgradeErrMsg, rollbackErrMsg))
	}

	//waiting for release to rollback
	success, waitErr := s.helmClient.WaitForReleaseRollback(kymahelm.NamespacedName{Namespace: s.component.Namespace, Name: s.component.GetReleaseName()})

	if waitErr != nil {
		return errors.New(fmt.Sprintf("%s\nHelm rollback of release \"%s\" failed with error: %s", upgradeErrMsg, s.component.GetReleaseName(), waitErr.Error()))
	} else {
		if success {
			return errors.New(fmt.Sprintf("%s\nHelm rollback of release \"%s\" was successfull", upgradeErrMsg, s.component.GetReleaseName()))
		} else {
			return errors.New(fmt.Sprintf("%s\nHelm rollback of release \"%s\" timed out", upgradeErrMsg, s.component.GetReleaseName()))
		}
	}
}

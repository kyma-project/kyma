package steps

import (
	"errors"
	"fmt"
	"log"
	"time"
)

type upgradeStep struct {
	installStep
	deployedRevision    int
	rollbackWaitTimeSec uint32
}

// Run method for upgradeStep triggers step upgrade via helm
func (s upgradeStep) Run() error {

	chartDir, err := s.sourceGetter.SrcDirFor(s.component)
	if err != nil {
		return err
	}

	releaseOverrides, releaseOverridesErr := s.overrideData.ForRelease(s.component.GetReleaseName())

	if releaseOverridesErr != nil {
		return releaseOverridesErr
	}

	upgradeResp, upgradeErr := s.helmClient.UpgradeRelease(
		chartDir,
		s.component.GetReleaseName(),
		releaseOverrides)

	if upgradeErr != nil {
		upgradeErrMsg := fmt.Sprintf("Helm upgrade error: %s", upgradeErr.Error())

		log.Println(fmt.Sprintf("Helm upgrade of %s failed. Performing rollback to last known deployed revision.", s.component.GetReleaseName()))
		_, err := s.helmClient.RollbackRelease(s.component.GetReleaseName(), 0)

		if err != nil {
			rollbackErrMsg := fmt.Sprintf("Helm rollback of %s failed with an error: %s", s.component.GetReleaseName(), err.Error())
			return errors.New(fmt.Sprintf("%s \n %s \n", upgradeErrMsg, rollbackErrMsg))
		}

		//waiting for release to rollback
		//TODO implement waiting method
		time.Sleep(time.Second * time.Duration(s.rollbackWaitTimeSec))

		return errors.New(fmt.Sprintf("%s\nHelm rollback of %s was successfull", upgradeErrMsg, s.component.GetReleaseName()))
	}

	s.helmClient.PrintRelease(upgradeResp)

	return nil
}

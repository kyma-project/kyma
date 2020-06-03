package steps

import (
	"errors"
	"fmt"
	"log"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
)

type installStep struct {
	step
	sourceGetter      kymasources.SourceGetter
	overrideData      overrides.OverrideData
	deleteWaitTimeSec uint32
}

// Run method for installStep triggers step installation via helm
func (s installStep) Run() error {

	chartDir, err := s.sourceGetter.SrcDirFor(s.component)
	if err != nil {
		return err
	}

	releaseOverrides, releaseOverridesErr := s.overrideData.ForRelease(s.component.GetReleaseName())

	if releaseOverridesErr != nil {
		return releaseOverridesErr
	}

	installResp, installErr := s.helmClient.InstallRelease(
		chartDir,
		s.component.Namespace,
		s.component.GetReleaseName(),
		releaseOverrides)

	if installErr != nil {
		return s.onError(installErr)
	}

	s.helmClient.PrintRelease(installResp.Release)

	return nil
}

func (s installStep) onError(installErr error) error {
	installErrMsg := fmt.Sprintf("Helm install error: %s", installErr.Error())

	isDeletable, checkErr := s.helmClient.IsReleaseDeletable(s.component.GetReleaseName())

	if checkErr != nil {
		statusErrMsg := fmt.Sprintf("Checking status of release \"%s\" failed with an error: %s", s.component.GetReleaseName(), checkErr.Error())
		return errors.New(fmt.Sprintf("%s \n %s \n", installErrMsg, statusErrMsg))
	}

	if isDeletable {

		log.Println(fmt.Sprintf("Helm installation of release \"%s\" failed. Deleting release before retrying.", s.component.GetReleaseName()))
		_, deleteErr := s.helmClient.DeleteRelease(s.component.GetReleaseName())

		if deleteErr != nil {
			deleteErrMsg := fmt.Sprintf("Helm delete of release \"%s\" failed with an error: %s", s.component.GetReleaseName(), deleteErr.Error())
			return errors.New(fmt.Sprintf("%s \n %s \n", installErrMsg, deleteErrMsg))
		}

		//waiting for release to be deleted
		success, waitErr := s.helmClient.WaitForReleaseDelete(s.component.GetReleaseName())
		if waitErr != nil {
			return errors.New(fmt.Sprintf("%s\nHelm delete of release \"%s\" failed with error: %s", installErrMsg, s.component.GetReleaseName(), waitErr.Error()))
		} else {
			if success {
				return errors.New(fmt.Sprintf("%s\nHelm delete of release \"%s\" was successfull", installErrMsg, s.component.GetReleaseName()))
			} else {
				return errors.New(fmt.Sprintf("%s\nHelm delete of release \"%s\" timed out", installErrMsg, s.component.GetReleaseName()))
			}
		}
	}

	return errors.New(installErrMsg)
}

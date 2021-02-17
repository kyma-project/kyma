package steps

import (
	"errors"
	"fmt"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
)

type installStep struct {
	step
	sourceGetter kymasources.SourceGetter
	overrideData overrides.OverrideData
}

// Run method for installStep triggers step installation via helm
func (s installStep) Run() error {

	chartDir, err := s.sourceGetter.SrcDirFor(s.component)
	if err != nil {
		return err
	}

	installResp, installErr := s.helmClient.InstallRelease(
		chartDir,
		s.GetNamespacedName(),
		s.overrideData.ForRelease(s.component.GetReleaseName()),
		string(s.step.profile))

	if installErr != nil {
		return s.onError(installErr)
	}

	s.helmClient.PrintRelease(installResp)

	return nil
}

func (s installStep) onError(installErr error) error {
	installErrMsg := fmt.Sprintf("Helm installation of release \"%s\" failed: %s", s.component.GetReleaseName(), installErr.Error())

	namespacedName := s.GetNamespacedName()

	isDeletable, checkErr := s.helmClient.IsReleaseDeletable(namespacedName)
	if checkErr != nil {
		statusErrMsg := fmt.Sprintf("Checking status of release \"%s\" failed with an error: %s", s.component.GetReleaseName(), checkErr.Error())
		return errors.New(fmt.Sprintf("%s \n %s \n", installErrMsg, statusErrMsg))
	}

	if isDeletable {

		installErrMsg = installErrMsg + "\n" + "Deleting release before retrying."

		if deleteErr := s.helmClient.UninstallRelease(namespacedName); deleteErr != nil {
			deleteErrMsg := fmt.Sprintf("Helm delete of release \"%s\" failed with an error: %s", s.component.GetReleaseName(), deleteErr.Error())
			return errors.New(fmt.Sprintf("%s \n %s \n", installErrMsg, deleteErrMsg))
		}

		//waiting for release to be deleted
		success, waitErr := s.helmClient.WaitForReleaseDelete(namespacedName)
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

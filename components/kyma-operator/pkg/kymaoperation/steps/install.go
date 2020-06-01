package steps

import (
	"errors"
	"fmt"
	"log"
	"time"

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

	installResp, installErr := s.helmClient.InstallRelease(
		chartDir,
		s.component.Namespace,
		s.component.GetReleaseName(),
		s.overrideData.ForRelease(s.component.GetReleaseName()))

	if installErr != nil {
		installErrMsg := fmt.Sprintf("Helm install error: %s", installErr.Error())

		isDeletable, err := s.helmClient.IsReleaseDeletable(s.component.Namespace, s.component.GetReleaseName())
		if err != nil {
			statusErrMsg := fmt.Sprintf("Checking status of %s failed with an error: %s", s.component.GetReleaseName(), err.Error())
			log.Println(statusErrMsg)
			return errors.New(fmt.Sprintf("%s \n %s \n", installErrMsg, statusErrMsg))
		}

		if isDeletable {

			log.Println(fmt.Sprintf("Helm installation of %s failed. Deleting before retrying installation.", s.component.GetReleaseName()))
			_, err := s.helmClient.DeleteRelease(s.component.Namespace, s.component.GetReleaseName())

			if err != nil {
				deleteErrMsg := fmt.Sprintf("Helm delete of %s failed with an error: %s", s.component.GetReleaseName(), err.Error())
				return errors.New(fmt.Sprintf("%s \n %s \n", installErrMsg, deleteErrMsg))
			}

			//waiting for release to be deleted
			//TODO implement waiting method
			time.Sleep(time.Second * time.Duration(s.deleteWaitTimeSec))

			return errors.New(fmt.Sprintf("%s\nHelm delete of %s was successfull", installErrMsg, s.component.GetReleaseName()))
		}

		return errors.New(installErrMsg)
	}

	s.helmClient.PrintRelease(installResp)

	return nil
}

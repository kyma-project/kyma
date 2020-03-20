package kymainstallation

import (
	"fmt"
	"log"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	"github.com/pkg/errors"
)

// Step defines the contract for a single installation/uninstallation operation
type Step interface {
	Run() error
	Status() (string, error)
	ToString() string
}

type step struct {
	helmClient kymahelm.ClientInterface
	component  v1alpha1.KymaComponent
}

// ToString method returns step details in readable string
func (s step) ToString() string {
	return fmt.Sprintf("Component: %s, Release: %s, Namespace: %s", s.component.Name, s.component.GetReleaseName(), s.component.Namespace)
}

// Status returns helm release status
func (s step) Status() (string, error) {
	return s.helmClient.ReleaseStatus(s.component.GetReleaseName())
}

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
		installErrMsg := fmt.Sprintf("Helm install error: %s ", installErr.Error())
		log.Println(installErrMsg)
		log.Println(fmt.Sprintf("Deleting release %s before retrying installation...", s.component.GetReleaseName()))
		_, err := s.helmClient.DeleteRelease(s.component.GetReleaseName())

		if err != nil {
			deleteErrMsg := fmt.Sprintf("Helm delete of %s failed with an error: %s", s.component.GetReleaseName(), err.Error())
			log.Println(deleteErrMsg)
			return errors.New(fmt.Sprintf("%s \n %s \n", installErrMsg, deleteErrMsg))
		}
		log.Println("Successfully deleted release")

		return errors.New(installErrMsg)
	}

	s.helmClient.PrintRelease(installResp.Release)

	return nil
}

type upgradeStep struct {
	installStep
}

// Run method for upgradeStep triggers step upgrade via helm
func (s upgradeStep) Run() error {

	//First get release history for possible future rollback
	releaseHistory, err := s.helmClient.ReleaseHistory(s.component.GetReleaseName(), 1)

	if err != nil {
		return err
	}

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
		releaseOverrides,
	)

	if upgradeErr != nil {
		log.Println("Helm upgrade error: " + upgradeErr.Error())
		upgradeFailedMsg := fmt.Sprintf("Helm upgrade error: %s", upgradeErr.Error())

		revision := releaseHistory.Releases[0].Version

		log.Println(fmt.Sprintf("Doing rollback of release %s to revision %d before retrying installation...", s.component.GetReleaseName(), revision))

		_, rollbackErr := s.helmClient.RollbackRelease(s.component.GetReleaseName(), revision)
		if rollbackErr != nil {
			rollbackFailedMsg := fmt.Sprintf("Helm rollback for release %s to revision %d failed with an error: %s ", s.component.GetReleaseName(), revision, rollbackErr.Error())
			log.Println(rollbackFailedMsg)
			return errors.New(fmt.Sprintf("%s \n %s \n", upgradeFailedMsg, rollbackFailedMsg))
		}
		log.Println("Successfully reverted release")

		return errors.New(upgradeFailedMsg)
	}
	s.helmClient.PrintRelease(upgradeResp.Release)

	return nil
}

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

type noStep struct {
	step
}

// Run method for noStep logs the information about missing release
func (s noStep) Run() error {
	log.Printf("Component %s is not deployed, skipping...", s.component.Name)
	return nil
}

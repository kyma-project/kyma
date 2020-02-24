package kymainstallation

import (
	"errors"
	"fmt"
	"log"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
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
		return errors.New("Helm install error: " + installErr.Error())
	}

	s.helmClient.PrintRelease(installResp.Release)

	return nil
}

type upgradeStep struct {
	installStep
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
		return errors.New("Helm upgrade error: " + upgradeErr.Error())
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

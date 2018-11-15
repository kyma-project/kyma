package kymainstallation

import (
	"errors"
	"fmt"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/installer/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/installer/pkg/kymasources"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

// Step represents contract for installation step
type Step interface {
	Run() error
	Status() (string, error)
	ToString() string
}

type step struct {
	helmClient   kymahelm.ClientInterface
	kymaPackage  kymasources.KymaPackage
	component    v1alpha1.KymaComponent
	overrideData overrides.OverrideData
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
}

// Run method for installStep triggers step installation via helm
func (s installStep) Run() error {
	chartDir := path.Join(s.kymaPackage.GetChartsDirPath(), s.component.Name)

	overrides, overridesErr := s.overrideData.ForRelease(s.component.GetReleaseName())

	if overridesErr != nil {
		return overridesErr
	}

	installResp, installErr := s.helmClient.InstallRelease(
		chartDir,
		s.component.Namespace,
		s.component.GetReleaseName(),
		overrides)

	if installErr != nil {
		return errors.New("Helm install error: " + installErr.Error())
	}

	s.helmClient.PrintRelease(installResp.Release)

	if s.component.Name == "core" {
		upgradeResp, upgradeErr := s.helmClient.UpgradeRelease(
			chartDir,
			s.component.GetReleaseName(),
			overrides)

		if upgradeErr != nil {
			return upgradeErr
		}

		s.helmClient.PrintRelease(upgradeResp.Release)
	}

	return nil
}

type upgradeStep struct {
	step
}

// Run method for upgradeStep triggers step upgrade via helm
func (s upgradeStep) Run() error {
	chartDir := path.Join(s.kymaPackage.GetChartsDirPath(), s.component.Name)

	overrides, overridesErr := s.overrideData.ForRelease(s.component.GetReleaseName())

	if overridesErr != nil {
		return overridesErr
	}

	upgradeResp, upgradeErr := s.helmClient.UpgradeRelease(
		chartDir,
		s.component.GetReleaseName(),
		overrides)

	if upgradeErr != nil {
		return errors.New("Helm upgrade error: " + upgradeErr.Error())
	}

	s.helmClient.PrintRelease(upgradeResp.Release)

	return nil
}

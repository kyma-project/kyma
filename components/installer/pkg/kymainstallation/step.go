package kymainstallation

import (
	"fmt"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/installer/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/installer/pkg/kymasources"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

// Step represents contract for installation step
type Step interface {
	Install() error
	Upgrade() error
	Status() (string, error)
	ToString() string
}

type step struct {
	helmClient              kymahelm.ClientInterface
	kymaPackage             kymasources.KymaPackage
	component               v1alpha1.KymaComponent
	legacyOverridesProvider overrides.LegacyProvider
}

// ToString method returns step details in readable string
func (s step) ToString() string {
	return fmt.Sprintf("Component: %s, Release: %s, Namespace: %s", s.component.Name, s.component.GetReleaseName(), s.component.Namespace)
}

// Install method triggers step installation via helm
func (s step) Install() error {
	chartDir := path.Join(s.kymaPackage.GetChartsDirPath(), s.component.Name)

	overrides, overridesErr := s.legacyOverridesProvider.GetForRelease(s.component)

	if overridesErr != nil {
		return overridesErr
	}

	installResp, installErr := s.helmClient.InstallRelease(
		chartDir,
		s.component.Namespace,
		s.component.GetReleaseName(),
		overrides)

	if installErr != nil {
		return installErr
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

// Upgrade method triggers step upgrade via helm
func (s step) Upgrade() error {
	chartDir := path.Join(s.kymaPackage.GetChartsDirPath(), s.component.Name)

	overrides, overridesErr := s.legacyOverridesProvider.GetForRelease(s.component)

	if overridesErr != nil {
		return overridesErr
	}

	upgradeResp, upgradeErr := s.helmClient.UpgradeRelease(
		chartDir,
		s.component.GetReleaseName(),
		overrides)

	if upgradeErr != nil {
		return upgradeErr
	}

	s.helmClient.PrintRelease(upgradeResp.Release)

	return nil
}

// Status returns helm release status
func (s step) Status() (string, error) {
	return s.helmClient.ReleaseStatus(s.component.GetReleaseName())
}

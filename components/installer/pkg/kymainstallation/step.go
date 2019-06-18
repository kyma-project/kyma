package kymainstallation

import (
	"errors"
	"fmt"
	"github.com/istio/pkg/log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/installer/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

// Step represents contract for installation step
type Step interface {
	Run() error
	Status() (string, error)
	ToString() string
}

type step struct {
	helmClient    kymahelm.ClientInterface
	chartsDirPath string
	component     v1alpha1.KymaComponent
	overrideData  overrides.OverrideData
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
	chartDir := path.Join(s.chartsDirPath, s.component.Name)

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
	step
}

// Run method for upgradeStep triggers step upgrade via helm
func (s upgradeStep) Run() error {
	chartDir := path.Join(s.chartsDirPath, s.component.Name)

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

// Run method for deleteStep triggers step delete via helm. Uninstall should not be retried, hence no error is returned.
func (s uninstallStep) Run() error {

	uninstallReleaseResponse, deleteErr := s.helmClient.DeleteRelease(s.component.GetReleaseName())

	if deleteErr != nil {
		log.Errorf("Helm delete error: %s", deleteErr.Error())
		return nil
	}

	s.helmClient.PrintRelease(uninstallReleaseResponse.Release)
	return nil
}

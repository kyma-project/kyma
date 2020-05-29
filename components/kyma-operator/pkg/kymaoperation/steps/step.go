package steps

import (
	"fmt"
	"log"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
)

// Step defines the contract for a single installation/uninstallation operation
// Installation step may be implemented an a Helm upgrade or install operation.
type Step interface {
	Run() error
	ToString() string
	GetReleaseName() string
}

// Helm functions common to all steps
type HelmClient interface {
	IsReleaseDeletable(rname string) (bool, error)
	InstallRelease(chartdir, ns, releasename, overrides string) (*kymahelm.Release, error)
	UpgradeRelease(chartDir, releaseName, overrides string) (*kymahelm.Release, error)
	RollbackRelease(releaseName string, revision int) (*kymahelm.Release, error)
	DeleteRelease(releaseName string) (*kymahelm.Release, error)
	PrintRelease(release *kymahelm.Release)
}

type step struct {
	helmClient HelmClient
	component  v1alpha1.KymaComponent
}

// ToString method returns step details in readable string
func (s step) ToString() string {
	return fmt.Sprintf("Component: %s, Release: %s, Namespace: %s", s.component.Name, s.component.GetReleaseName(), s.component.Namespace)
}

func (s step) GetReleaseName() string {
	return s.component.GetReleaseName()
}

// Used to support the contract in case there's no action to do (e.g. uninstalling a non-existing release)
type noStep struct {
	step
}

// Run method for noStep logs the information about missing release
func (s noStep) Run() error {
	log.Printf("Component %s is not deployed, skipping...", s.component.Name)
	return nil
}

package kymainstallation

import (
	"fmt"
	"log"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

// Step defines the contract for a single installation/uninstallation operation
// Installation step may be implemented an a Helm upgrade or install operation.
type Step interface {
	Run() error
	ToString() string
}

// Helm methods common to all steps
type HelmClient interface {
	IsReleaseDeletable(rname string) (bool, error)
	InstallRelease(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error)
	UpgradeRelease(chartDir, releaseName, overrides string) (*rls.UpdateReleaseResponse, error)
	RollbackRelease(releaseName string, revision int32) (*rls.RollbackReleaseResponse, error)
	DeleteRelease(releaseName string) (*rls.UninstallReleaseResponse, error)
	PrintRelease(release *release.Release)
}

type step struct {
	helmClient HelmClient
	component  v1alpha1.KymaComponent
}

// ToString method returns step details in readable string
func (s step) ToString() string {
	return fmt.Sprintf("Component: %s, Release: %s, Namespace: %s", s.component.Name, s.component.GetReleaseName(), s.component.Namespace)
}

type noStep struct {
	step
}

// Run method for noStep logs the information about missing release
func (s noStep) Run() error {
	log.Printf("Component %s is not deployed, skipping...", s.component.Name)
	return nil
}

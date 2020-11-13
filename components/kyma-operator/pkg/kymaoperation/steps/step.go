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
	String() string
	GetReleaseName() string
	GetNamespacedName() kymahelm.NamespacedName
}

type step struct {
	helmClient kymahelm.ClientInterface
	component  v1alpha1.KymaComponent
	profile    v1alpha1.KymaProfile
}

// ToString method returns step details in readable string
func (s step) String() string {
	return fmt.Sprintf("Component: %s, Release: %s, Namespace: %s", s.component.Name, s.component.GetReleaseName(), s.component.Namespace)
}

func (s step) GetReleaseName() string {
	return s.component.GetReleaseName()
}

func (s step) GetNamespacedName() kymahelm.NamespacedName {
	return kymahelm.NamespacedName{Name: s.GetReleaseName(), Namespace: s.component.Namespace}
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

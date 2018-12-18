package access

import (
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/application-broker/internal"
)

type instanceFinder interface {
	FindOne(m func(i *internal.Instance) bool) (*internal.Instance, error)
}

// NewUniquenessProvisionChecker creates new access checker
func NewUniquenessProvisionChecker(iFind instanceFinder) *UniquenessProvisionChecker {
	return &UniquenessProvisionChecker{
		InstanceFinder: iFind,
	}
}

// UniquenessProvisionChecker is a checker which ensures that only one instance of a specific class is in namespace.
type UniquenessProvisionChecker struct {
	InstanceFinder instanceFinder
}

// CanProvision performs actual check
func (c *UniquenessProvisionChecker) CanProvision(iID internal.InstanceID, rsID internal.RemoteServiceID, ns internal.Namespace) (CanProvisionOutput, error) {
	i, err := c.InstanceFinder.FindOne(func(i *internal.Instance) bool {
		// exclude itself
		if i.ID == iID {
			return false
		}
		if i.Namespace != ns {
			return false
		}
		if i.ServiceID != internal.ServiceID(rsID) {
			return false
		}
		if i.State == internal.InstanceStateFailed {
			return false
		}
		return true
	})

	if err != nil {
		return CanProvisionOutput{}, errors.Wrapf(err, "while finding instance")
	}
	if i == nil {
		return c.responseAllow(), nil
	}

	return c.responseDeny(), nil
}

func (c *UniquenessProvisionChecker) responseAllow() CanProvisionOutput {
	return CanProvisionOutput{Allowed: true}
}

func (c *UniquenessProvisionChecker) responseDeny() CanProvisionOutput {
	return CanProvisionOutput{Allowed: false, Reason: "already activated"}
}

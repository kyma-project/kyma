package access

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	versioned "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
)

//go:generate mockery -name=ProvisionChecker -output=automock -outpkg=automock -case=underscore

// ProvisionChecker define methods for checking if provision can succeed
type ProvisionChecker interface {
	CanProvision(iID internal.InstanceID, rsID internal.ApplicationServiceID, namespace internal.Namespace, maxWaitTime time.Duration) (CanProvisionOutput, error)
}

// CanProvisionOutput aggregates following information: if provision can be performed and reason
type CanProvisionOutput struct {
	Allowed bool
	Reason  string
}

// New creates new aggregated checker
func New(appFinder applicationFinder, mappingClient versioned.ApplicationconnectorV1alpha1Interface, iFind instanceFinder, apiPackagesSupport bool) *AggregatedChecker {
	return &AggregatedChecker{
		mappingExistsProvisionChecker: NewMappingExistsProvisionChecker(appFinder, mappingClient),
		uniquenessProvisionChecker:    NewUniquenessProvisionChecker(iFind),
		apiPackagesSupport:            apiPackagesSupport,
	}
}

// AggregatedChecker is a checker which aggregates multiple checks.
// All checks are performed sequentially. First failed check stops further ones.
type AggregatedChecker struct {
	mappingExistsProvisionChecker interface {
		CanProvision(rsID internal.ApplicationServiceID, ns internal.Namespace, maxWaitTime time.Duration) (CanProvisionOutput, error)
	}
	uniquenessProvisionChecker interface {
		CanProvision(iID internal.InstanceID, rsID internal.ApplicationServiceID, ns internal.Namespace) (CanProvisionOutput, error)
	}
	apiPackagesSupport bool
}

// CanProvision performs actual check
func (c *AggregatedChecker) CanProvision(iID internal.InstanceID, rsID internal.ApplicationServiceID, ns internal.Namespace, maxWaitTime time.Duration) (CanProvisionOutput, error) {
	res, err := c.mappingExistsProvisionChecker.CanProvision(rsID, ns, maxWaitTime)
	if err != nil {
		return CanProvisionOutput{}, errors.Wrap(err, "while calling mappingExistsProvisionChecker")
	}
	if c.apiPackagesSupport || !res.Allowed {
		return res, nil
	}

	res, err = c.uniquenessProvisionChecker.CanProvision(iID, rsID, ns)
	if err != nil {
		return CanProvisionOutput{}, errors.Wrap(err, "while calling uniquenessProvisionChecker")
	}

	return res, nil
}

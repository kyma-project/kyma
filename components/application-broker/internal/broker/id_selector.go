package broker

import (
	"github.com/kyma-project/kyma/components/application-broker/internal"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type IDSelector struct {
	apiPackagesSupport bool
}

func NewIDSelector(apiPackagesSupport bool) *IDSelector {
	return &IDSelector{apiPackagesSupport: apiPackagesSupport}
}

func (s *IDSelector) SelectID(req interface{}) internal.ApplicationServiceID {
	var svcID, planID string
	switch d := req.(type) {
	case *osb.BindRequest:
		svcID, planID = d.ServiceID, d.PlanID
	case *osb.ProvisionRequest:
		svcID, planID = d.ServiceID, d.PlanID
	case *osb.DeprovisionRequest:
		svcID, planID = d.ServiceID, d.PlanID
	}

	return s.SelectApplicationServiceID(svcID, planID)
}

func (s *IDSelector) SelectApplicationServiceID(serviceID, planID string) internal.ApplicationServiceID {
	// In new approach ApplicationServiceID == req Plan ID
	if s.apiPackagesSupport {
		return internal.ApplicationServiceID(planID)
	}

	// In old approach ApplicationServiceID == req Service ID == Class ID
	return internal.ApplicationServiceID(serviceID)
}

package broker

import (
	"testing"

	osb "github.com/kubernetes-sigs/go-open-service-broker-client/v2"

	"github.com/kyma-project/kyma/components/application-broker/internal"

	"github.com/stretchr/testify/assert"
)

func TestIDSelector_SelectApplicationServiceID(t *testing.T) {
	for name, tc := range map[string]struct {
		apiPackagesSupport bool
		serviceID          string
		planID             string
		want               internal.ApplicationServiceID
	}{
		"support api package": {
			apiPackagesSupport: true,
			serviceID:          "1234",
			planID:             "1234-plan",
			want:               "1234-plan",
		},
		"not support api package": {
			apiPackagesSupport: false,
			serviceID:          "1234",
			planID:             "1234-plan",
			want:               "1234",
		},
	} {
		t.Run(name, func(t *testing.T) {
			// Given
			s := NewIDSelector(tc.apiPackagesSupport)

			// When
			appID := s.SelectApplicationServiceID(tc.serviceID, tc.planID)

			// Then
			assert.Equal(t, tc.want, appID)
		})
	}
}

func TestIDSelector_SelectID(t *testing.T) {
	for name, tc := range map[string]struct {
		apiPackagesSupport bool
		req                *osb.ProvisionRequest
		want               internal.ApplicationServiceID
	}{
		"support api package": {
			apiPackagesSupport: true,
			req: &osb.ProvisionRequest{
				ServiceID: "7890",
				PlanID:    "7890-plan",
			},
			want: "7890-plan",
		},
		"not support api package": {
			apiPackagesSupport: false,
			req: &osb.ProvisionRequest{
				ServiceID: "7890",
				PlanID:    "7890-plan",
			},
			want: "7890",
		},
	} {
		t.Run(name, func(t *testing.T) {
			// Given
			s := NewIDSelector(tc.apiPackagesSupport)

			// When
			appID := s.SelectID(tc.req)

			// Then
			assert.Equal(t, tc.want, appID)
		})
	}
}

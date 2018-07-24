package status

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestInstanceExtractor_Status(t *testing.T) {
	ext := InstanceExtractor{}
	for tn, tc := range map[string]struct {
		given    []v1beta1.ServiceInstanceCondition
		expected ServiceInstanceStatus
	}{
		"ReadyStatus": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionTrue,
				}, {
					Type:   v1beta1.ServiceInstanceConditionFailed,
					Status: v1beta1.ConditionFalse,
				},
			},
			expected: ServiceInstanceStatus{
				Type: ServiceInstanceStatusTypeRunning,
			},
		},
		"FailedStatus": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
				}, {
					Type:   v1beta1.ServiceInstanceConditionFailed,
					Status: v1beta1.ConditionTrue,
				},
			},
			expected: ServiceInstanceStatus{
				Type: ServiceInstanceStatusTypeFailed,
			},
		},
		"FailedStatusv2": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
					Reason: "ProvisionCallFailed",
				},
			},
			expected: ServiceInstanceStatus{
				Type:   ServiceInstanceStatusTypeFailed,
				Reason: "ProvisionCallFailed",
			},
		},
		"FailedStatusv3": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
					Reason: "ReferencesNonexistentServiceClass",
				},
			},
			expected: ServiceInstanceStatus{
				Type:   ServiceInstanceStatusTypeFailed,
				Reason: "ReferencesNonexistentServiceClass",
			},
		},
		"FailedStatusv4": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
					Reason: "ErrorFindingNamespaceForInstance",
				},
			},
			expected: ServiceInstanceStatus{
				Type:   ServiceInstanceStatusTypeFailed,
				Reason: "ErrorFindingNamespaceForInstance",
			},
		},
		"FailedStatusv5": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
					Reason: "InvalidDeprovisionStatus",
				},
			},
			expected: ServiceInstanceStatus{
				Type:   ServiceInstanceStatusTypeFailed,
				Reason: "InvalidDeprovisionStatus",
			},
		},
		"PendingStatus": {
			given: []v1beta1.ServiceInstanceCondition{},
			expected: ServiceInstanceStatus{
				Type: ServiceInstanceStatusTypePending,
			},
		},
		"PendingStatusv2": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
					Reason: "UnknownReason",
				},
			},
			expected: ServiceInstanceStatus{
				Type:   ServiceInstanceStatusTypePending,
				Reason: "UnknownReason",
			},
		},
		"ProvisioningStatus": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
					Reason: "Provisioning",
				},
			},
			expected: ServiceInstanceStatus{
				Type:   ServiceInstanceStatusTypeProvisioning,
				Reason: "Provisioning",
			},
		},
		"ProvisioningStatusv2": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
					Reason: "ProvisionRequestInFlight",
				},
			},
			expected: ServiceInstanceStatus{
				Type:   ServiceInstanceStatusTypeProvisioning,
				Reason: "ProvisionRequestInFlight",
			},
		},
		"ProvisioningStatusv3": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
					Reason: "UpdatingInstance",
				},
			},
			expected: ServiceInstanceStatus{
				Type:   ServiceInstanceStatusTypeProvisioning,
				Reason: "UpdatingInstance",
			},
		},
		"ProvisioningStatusv4": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
					Reason: "UpdateInstanceRequestInFlight",
				},
			},
			expected: ServiceInstanceStatus{
				Type:   ServiceInstanceStatusTypeProvisioning,
				Reason: "UpdateInstanceRequestInFlight",
			},
		},
		"DeprovisioningStatus": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
					Reason: "Deprovisioning",
				},
			},
			expected: ServiceInstanceStatus{
				Type:   ServiceInstanceStatusTypeDeprovisioning,
				Reason: "Deprovisioning",
			},
		},
		"DeprovisioningStatusv2": {
			given: []v1beta1.ServiceInstanceCondition{
				{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionFalse,
					Reason: "DeprovisionRequestInFlight",
				},
			},
			expected: ServiceInstanceStatus{
				Type:   ServiceInstanceStatusTypeDeprovisioning,
				Reason: "DeprovisionRequestInFlight",
			},
		},
	} {
		instance := &v1beta1.ServiceInstance{
			Status: v1beta1.ServiceInstanceStatus{
				AsyncOpInProgress: false,
				Conditions:        tc.given,
			},
		}
		t.Run(tn, func(t *testing.T) {
			result := ext.Status(instance)
			assert.Equal(t, &tc.expected, result)
		})
	}
}

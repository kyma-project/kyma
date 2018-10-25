package status

import (
	sbuTypes "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// -- Reasons which can occur during add reconciliation of ServiceBindingUsage

	// ServiceBindingGetErrorReason is added in an usage when we cannot get a given ServiceBinding
	ServiceBindingGetErrorReason = "ServiceBindingGetError"
	// ServiceBindingOngoingAsyncOptReason is added in an usage when given ServiceBinding has ongoing async operation
	ServiceBindingOngoingAsyncOptReason = "ServiceBindingAsyncOperationInProgressError"
	// ServiceBindingNotReadyReason id added in an usage when given ServiceBinding is not in ready state
	ServiceBindingNotReadyReason = "ServiceBindingNotReadyError"
	// PodPresetUpsertErrorReason is added in an usage when we cannot create a new PodPreset
	PodPresetUpsertErrorReason = "PodPresetUpsertError"
	// FetchBindingLabelsErrorReason is added in an usage when we cannot fetch labels from given ClusterServiceClass or ServiceClass
	FetchBindingLabelsErrorReason = "ServiceClassGetBindingLabelsError"
	// ApplyLabelsConflictErrorReason is added in a usage when we cannot add labels to the given resource because they already exists
	ApplyLabelsConflictErrorReason = "ApplyLabelsConflictError"
	// EnsureLabelsAppliedErrorReason is added in a usage when we cannot add labels to the given resource, e.g. given resource does not exits
	EnsureLabelsAppliedErrorReason = "EnsureLabelsAppliedError"
	// AddOwnerReferenceErrorReason is added in a usage when we cannot add an OwnerReference to the given ServiceBinding
	AddOwnerReferenceErrorReason = "AddOwnerReferenceError"

	// -- Reasons which can occur during delete reconciliation of ServiceBindingUsage

	// EnsureLabelsDeletedErrorReason is added in a usage when we cannot deleted labels from the given resource
	EnsureLabelsDeletedErrorReason = "EnsureLabelsDeletedError"
	// PodPresetDeleteErrorReason is added in an usage when we cannot delete a new PodPreset
	PodPresetDeleteErrorReason = "PodPresetDeleteError"
	// GetStoredSpecError is added in an usage when we cannot get stored spec for given ServiceBindingUsage
	GetStoredSpecError = "GetStoredSBUSpecError"
)

// TimeNowFn is used for getting time for a new condition.
// It's exported to allow client to mock it in tests.
var TimeNowFn = metaV1.Now

// NewUsageCondition creates a new usage condition.
func NewUsageCondition(condType sbuTypes.ServiceBindingUsageConditionType, status sbuTypes.ConditionStatus, reason, message string) *sbuTypes.ServiceBindingUsageCondition {
	return &sbuTypes.ServiceBindingUsageCondition{
		Type:               condType,
		Status:             status,
		LastUpdateTime:     TimeNowFn(),
		LastTransitionTime: TimeNowFn(),
		Reason:             reason,
		Message:            message,
	}
}

// GetUsageCondition returns the condition with the provided type or nil if not found.
func GetUsageCondition(status sbuTypes.ServiceBindingUsageStatus, condType sbuTypes.ServiceBindingUsageConditionType) *sbuTypes.ServiceBindingUsageCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// SetUsageCondition updates the usage to include the provided condition.
// If the condition that we are about to add already exists and has the same status then
// we are not going to update the LastTransitionTime.
func SetUsageCondition(status *sbuTypes.ServiceBindingUsageStatus, condition sbuTypes.ServiceBindingUsageCondition) {
	currentCond := GetUsageCondition(*status, condition.Type)

	// Do not update lastTransitionTime if the status of the condition doesn't change
	if currentCond != nil && currentCond.Status == condition.Status {
		condition.LastTransitionTime = currentCond.LastTransitionTime
	}
	newConditions := filterOutCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, condition)
}

// filterOutCondition returns a new slice without conditions with the given type
func filterOutCondition(conditions []sbuTypes.ServiceBindingUsageCondition, condType sbuTypes.ServiceBindingUsageConditionType) []sbuTypes.ServiceBindingUsageCondition {
	var newConditions []sbuTypes.ServiceBindingUsageCondition
	for _, c := range conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}

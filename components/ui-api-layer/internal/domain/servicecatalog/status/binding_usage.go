package status

import (
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

type BindingUsageExtractor struct{}

func (ext *BindingUsageExtractor) Status(conditions []v1alpha1.ServiceBindingUsageCondition) gqlschema.ServiceBindingUsageStatus {
	activeConditions := ext.findActiveConditions(conditions)

	if len(conditions) == 0 {
		return gqlschema.ServiceBindingUsageStatus{
			Type: gqlschema.ServiceBindingUsageStatusTypePending,
		}
	}
	if cond, found := ext.findReadyCondition(activeConditions); found {
		return gqlschema.ServiceBindingUsageStatus{
			Type:    gqlschema.ServiceBindingUsageStatusTypeReady,
			Message: cond.Message,
			Reason:  cond.Reason,
		}
	}
	if cond, found := ext.findReadyCondition(conditions); found {
		return gqlschema.ServiceBindingUsageStatus{
			Type:    gqlschema.ServiceBindingUsageStatusTypeFailed,
			Message: cond.Message,
			Reason:  cond.Reason,
		}
	}

	return gqlschema.ServiceBindingUsageStatus{
		Type: gqlschema.ServiceBindingUsageStatusTypeUnknown,
	}
}

func (*BindingUsageExtractor) findActiveConditions(conditions []v1alpha1.ServiceBindingUsageCondition) []v1alpha1.ServiceBindingUsageCondition {
	var result []v1alpha1.ServiceBindingUsageCondition
	for _, cond := range conditions {
		if cond.Status == v1alpha1.ConditionTrue {
			result = append(result, cond)
		}
	}
	return result
}

func (*BindingUsageExtractor) findReadyCondition(conditions []v1alpha1.ServiceBindingUsageCondition) (v1alpha1.ServiceBindingUsageCondition, bool) {
	for _, item := range conditions {
		if item.Type == v1alpha1.ServiceBindingUsageReady {
			return item, true
		}
	}
	return v1alpha1.ServiceBindingUsageCondition{}, false
}

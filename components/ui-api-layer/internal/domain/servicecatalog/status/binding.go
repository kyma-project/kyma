package status

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

type BindingExtractor struct{}

func (ext *BindingExtractor) Status(conditions []v1beta1.ServiceBindingCondition) gqlschema.ServiceBindingStatus {
	activeConditions := ext.findActiveConditions(conditions)

	if len(conditions) == 0 {
		return gqlschema.ServiceBindingStatus{
			Type: gqlschema.ServiceBindingStatusTypePending,
		}
	}
	if cond, found := ext.findReadyCondition(activeConditions); found {
		return gqlschema.ServiceBindingStatus{
			Type:    gqlschema.ServiceBindingStatusTypeReady,
			Message: cond.Message,
			Reason:  cond.Reason,
		}
	}
	if cond, found := ext.findReadyCondition(conditions); found {
		return gqlschema.ServiceBindingStatus{
			Type:    gqlschema.ServiceBindingStatusTypeFailed,
			Message: cond.Message,
			Reason:  cond.Reason,
		}
	}
	if cond, found := ext.findFailedCondition(activeConditions); found {
		return gqlschema.ServiceBindingStatus{
			Type:    gqlschema.ServiceBindingStatusTypeFailed,
			Message: cond.Message,
			Reason:  cond.Reason,
		}
	}

	return gqlschema.ServiceBindingStatus{
		Type: gqlschema.ServiceBindingStatusTypeUnknown,
	}
}

func (*BindingExtractor) findActiveConditions(conditions []v1beta1.ServiceBindingCondition) []v1beta1.ServiceBindingCondition {
	var result []v1beta1.ServiceBindingCondition
	for _, cond := range conditions {
		if cond.Status == v1beta1.ConditionTrue {
			result = append(result, cond)
		}
	}
	return result
}

func (*BindingExtractor) findReadyCondition(conditions []v1beta1.ServiceBindingCondition) (v1beta1.ServiceBindingCondition, bool) {
	for _, item := range conditions {
		if item.Type == v1beta1.ServiceBindingConditionReady {
			return item, true
		}
	}
	return v1beta1.ServiceBindingCondition{}, false
}

func (*BindingExtractor) findFailedCondition(conditions []v1beta1.ServiceBindingCondition) (v1beta1.ServiceBindingCondition, bool) {
	for _, item := range conditions {
		if item.Type == v1beta1.ServiceBindingConditionFailed {
			return item, true
		}
	}
	return v1beta1.ServiceBindingCondition{}, false
}

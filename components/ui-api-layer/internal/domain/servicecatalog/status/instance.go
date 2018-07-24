package status

import (
	"strings"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

type InstanceExtractor struct{}

type ServiceInstanceStatusType string

const (
	ServiceInstanceStatusTypeRunning        ServiceInstanceStatusType = "RUNNING"
	ServiceInstanceStatusTypeProvisioning   ServiceInstanceStatusType = "PROVISIONING"
	ServiceInstanceStatusTypeDeprovisioning ServiceInstanceStatusType = "DEPROVISIONING"
	ServiceInstanceStatusTypePending        ServiceInstanceStatusType = "PENDING"
	ServiceInstanceStatusTypeFailed         ServiceInstanceStatusType = "FAILED"
)

type ServiceInstanceStatus struct {
	Type    ServiceInstanceStatusType
	Reason  string
	Message string
}

func (ext *InstanceExtractor) Status(serviceInstance *v1beta1.ServiceInstance) *ServiceInstanceStatus {
	if serviceInstance == nil {
		return nil
	}

	conditions := serviceInstance.Status.Conditions
	activeCondition := ext.findActiveConditions(conditions)

	if len(conditions) == 0 {
		return &ServiceInstanceStatus{
			Type: ServiceInstanceStatusTypePending,
		}
	}
	if cond, found := ext.findConditionOfType(activeCondition, v1beta1.ServiceInstanceConditionReady); found {
		return &ServiceInstanceStatus{
			Type:    ServiceInstanceStatusTypeRunning,
			Reason:  cond.Reason,
			Message: cond.Message,
		}
	}
	if cond, found := ext.findConditionOfType(activeCondition, v1beta1.ServiceInstanceConditionFailed); found {
		return &ServiceInstanceStatus{
			Type:    ServiceInstanceStatusTypeFailed,
			Reason:  cond.Reason,
			Message: cond.Message,
		}
	}

	condition, _ := ext.findConditionOfType(conditions, v1beta1.ServiceInstanceConditionReady)
	return &ServiceInstanceStatus{
		Type:    ext.getProvisioningStatus(condition.Reason),
		Message: condition.Message,
		Reason:  condition.Reason,
	}
}

func (*InstanceExtractor) findActiveConditions(conditions []v1beta1.ServiceInstanceCondition) []v1beta1.ServiceInstanceCondition {
	var result []v1beta1.ServiceInstanceCondition
	for _, cond := range conditions {
		if cond.Status == v1beta1.ConditionTrue {
			result = append(result, cond)
		}
	}
	return result
}

func (*InstanceExtractor) findConditionOfType(conditions []v1beta1.ServiceInstanceCondition, typeOf v1beta1.ServiceInstanceConditionType) (v1beta1.ServiceInstanceCondition, bool) {
	for _, cond := range conditions {
		if cond.Type == typeOf {
			return cond, true
		}
	}
	return v1beta1.ServiceInstanceCondition{}, false
}

func (ext *InstanceExtractor) getProvisioningStatus(reason string) ServiceInstanceStatusType {
	failedStatus := []string{"Error", "Nonexistent", "Failed", "Deleted", "Invalid"}
	provisionedStatus := []string{"Provision", "Updat"}
	deprovisionedStatus := []string{"Deprovision"}

	switch {
	case ext.containsReason(reason, failedStatus):
		return ServiceInstanceStatusTypeFailed
	case ext.containsReason(reason, provisionedStatus):
		return ServiceInstanceStatusTypeProvisioning
	case ext.containsReason(reason, deprovisionedStatus):
		return ServiceInstanceStatusTypeDeprovisioning
	default:
		return ServiceInstanceStatusTypePending
	}
}

func (*InstanceExtractor) containsReason(reason string, subStrings []string) bool {
	for _, subString := range subStrings {
		if strings.Contains(reason, subString) {
			return true
		}
	}
	return false
}

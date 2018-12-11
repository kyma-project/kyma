package controller

import (
	"fmt"

	sbuTypes "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
)

type bindingUsageStatus struct {
	sbuType   sbuTypes.ServiceBindingUsageConditionType
	condition sbuTypes.ConditionStatus
	reason    string
	message   string
}

func newBindingUsageStatus() *bindingUsageStatus {
	return &bindingUsageStatus{
		sbuType: sbuTypes.ServiceBindingUsageReady,
		reason:  "",
		message: "",
	}
}

func (s *bindingUsageStatus) wrapMessageForFailed(msg string) {
	if s.condition == sbuTypes.ConditionTrue {
		return
	}

	s.message = fmt.Sprintf("%s; %s", msg, s.message)
}

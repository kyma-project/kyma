package status_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/status"
	sbuTypes "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	condReady = func() *sbuTypes.ServiceBindingUsageCondition {
		return status.NewUsageCondition(
			sbuTypes.ServiceBindingUsageReady, sbuTypes.ConditionTrue,
			"", "")
	}

	condNotReady = func() *sbuTypes.ServiceBindingUsageCondition {
		return status.NewUsageCondition(
			sbuTypes.ServiceBindingUsageReady, sbuTypes.ConditionTrue,
			"NotSoAwesome", "Failed")
	}

	statusWithoutCond = func() sbuTypes.ServiceBindingUsageStatus {
		return sbuTypes.ServiceBindingUsageStatus{
			Conditions: []sbuTypes.ServiceBindingUsageCondition{},
		}
	}

	statusWithCond = func(cond *sbuTypes.ServiceBindingUsageCondition) sbuTypes.ServiceBindingUsageStatus {
		return sbuTypes.ServiceBindingUsageStatus{
			Conditions: []sbuTypes.ServiceBindingUsageCondition{*cond},
		}
	}
)

func TestGetUsageCondition(t *testing.T) {
	tests := map[string]struct {
		givenStatus sbuTypes.ServiceBindingUsageStatus

		condExpected bool
	}{
		"condition exists": {
			givenStatus: statusWithCond(condReady()),

			condExpected: true,
		},
		"condition does not exist": {
			givenStatus: statusWithoutCond(),

			condExpected: false,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			cond := status.GetUsageCondition(tc.givenStatus, sbuTypes.ServiceBindingUsageReady)

			// then
			exists := cond != nil
			assert.Equal(t, tc.condExpected, exists)
		})
	}
}

func TestSetUsageConditionSetCondForTheFirstTime(t *testing.T) {
	// given
	var (
		usageStatus    = statusWithoutCond()
		givenCond      = condReady()
		expectedStatus = statusWithCond(givenCond)
	)
	// when
	status.SetUsageCondition(&usageStatus, *givenCond)

	// then
	assert.Equal(t, usageStatus, expectedStatus)
}

func TestSetUsageConditionSetCondWhichAlreadyExistsInStatus(t *testing.T) {
	// given
	fixCond := condNotReady()
	usageStatus := statusWithCond(fixCond)

	newCond := fixCond.DeepCopy()
	newCond.Message = "Still failing"
	newCond.LastUpdateTime = metaV1.NewTime(fixCond.LastUpdateTime.Add(time.Hour))
	newCond.LastTransitionTime = metaV1.NewTime(fixCond.LastTransitionTime.Add(2 * time.Hour))

	// when
	status.SetUsageCondition(&usageStatus, *newCond)

	// then
	require.Len(t, usageStatus.Conditions, 1)

	assert.Equal(t, newCond.Status, usageStatus.Conditions[0].Status)
	assert.Equal(t, newCond.Reason, usageStatus.Conditions[0].Reason)
	assert.Equal(t, newCond.Message, usageStatus.Conditions[0].Message)
	assert.Equal(t, newCond.LastUpdateTime, usageStatus.Conditions[0].LastUpdateTime)

	// LastTransitionTime should not be update by set method, when cond already exists and has the same status
	assert.NotEqual(t, newCond.LastTransitionTime, usageStatus.Conditions[0].LastTransitionTime)
	assert.Equal(t, fixCond.LastTransitionTime, usageStatus.Conditions[0].LastTransitionTime)

}

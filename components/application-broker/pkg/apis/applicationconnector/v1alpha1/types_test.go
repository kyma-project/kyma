package v1alpha1_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestApplicationMappingEmptySpec(t *testing.T) {
	// given
	am := v1alpha1.ApplicationMapping{
		Spec: v1alpha1.ApplicationMappingSpec{},
	}

	// when/then
	assert.True(t, am.IsAllApplicationServicesEnabled())
	assert.True(t, am.IsServiceEnabled("id-001"))
}

func TestApplicationMappingEmptyServiceList(t *testing.T) {
	// given
	am := v1alpha1.ApplicationMapping{
		Spec: v1alpha1.ApplicationMappingSpec{
			Services: []v1alpha1.ApplicationMappingService{},
		},
	}

	// when/then
	assert.False(t, am.IsAllApplicationServicesEnabled())
	assert.False(t, am.IsServiceEnabled("id-001"))
}

func TestApplicationMappingTwoServices(t *testing.T) {
	// given
	am := v1alpha1.ApplicationMapping{
		Spec: v1alpha1.ApplicationMappingSpec{
			Services: []v1alpha1.ApplicationMappingService{{ID: "id-001"}, {ID: "id-002"}},
		},
	}

	// when/then
	assert.False(t, am.IsAllApplicationServicesEnabled())
	assert.True(t, am.IsServiceEnabled("id-001"))
	assert.True(t, am.IsServiceEnabled("id-002"))
	assert.False(t, am.IsServiceEnabled("id-003"))
}

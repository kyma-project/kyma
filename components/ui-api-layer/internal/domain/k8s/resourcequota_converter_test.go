package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResourceQuotaConverter_ToGQLs(t *testing.T) {
	// GIVEN
	converter := &resourceQuotaConverter{}

	// WHEN
	result := converter.ToGQLs([]*v1.ResourceQuota{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mem-default",
				Namespace: "production",
			},
		},
	})

	// THEN
	assert.Equal(t, []gqlschema.ResourceQuota{
		{Name: "mem-default"},
	}, result)
}

func TestResourceQuotaConverter_ToGQL(t *testing.T) {
	for tn, tc := range map[string]struct {
		given    *v1.ResourceQuota
		expected *gqlschema.ResourceQuota
	}{
		"empty": {
			given: &v1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mem-default",
					Namespace: "production",
				},
			},
			expected: &gqlschema.ResourceQuota{
				Name: "mem-default",
			},
		},
		"full": {
			given: &v1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mem-default",
					Namespace: "production",
				},
				Spec: v1.ResourceQuotaSpec{
					Hard: v1.ResourceList{
						v1.ResourcePods:           resource.MustParse("10"),
						v1.ResourceLimitsCPU:      resource.MustParse("900m"),
						v1.ResourceLimitsMemory:   resource.MustParse("1Gi"),
						v1.ResourceRequestsCPU:    resource.MustParse("500m"),
						v1.ResourceRequestsMemory: resource.MustParse("512Mi"),
					},
				},
			},
			expected: &gqlschema.ResourceQuota{
				Name: "mem-default",
				Pods: ptrStr("10"),
				Limits: gqlschema.ResourceValues{
					Cpu:    ptrStr("900m"),
					Memory: ptrStr("1Gi"),
				},
				Requests: gqlschema.ResourceValues{
					Cpu:    ptrStr("500m"),
					Memory: ptrStr("512Mi"),
				},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			converter := &resourceQuotaConverter{}

			// WHEN
			result := converter.ToGQL(tc.given)

			// THEN
			assert.Equal(t, tc.expected, result)
		})

	}
}

func ptrStr(str string) *string {
	return &str
}

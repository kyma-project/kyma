package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLimitRangeConverter_ToGQL(t *testing.T) {
	for tn, tc := range map[string]struct {
		given    *v1.LimitRange
		expected *gqlschema.LimitRange
	}{
		"empty": {
			given: &v1.LimitRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fixLimitRangeName(),
					Namespace: fixLimitRangeNamespace(),
				},
			},
			expected: &gqlschema.LimitRange{
				Name:   fixLimitRangeName(),
				Limits: make([]gqlschema.LimitRangeItem, 0),
			},
		},
		"full": {
			given:    fixLimitRange(),
			expected: fixGQLLimitRange(),
		},
		"nil": {
			given: &v1.LimitRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fixLimitRangeName(),
					Namespace: fixLimitRangeNamespace(),
				},
				Spec: v1.LimitRangeSpec{
					Limits: []v1.LimitRangeItem{
						{
							Max: v1.ResourceList{},
						},
					},
				},
			},
			expected: &gqlschema.LimitRange{
				Name: fixLimitRangeName(),
				Limits: []gqlschema.LimitRangeItem{
					{
						Max: gqlschema.ResourceType{
							Memory: nil,
							Cpu:    nil,
						},
					},
				},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			conv := limitRangeConverter{}

			// WHEN
			gql := conv.ToGQL(tc.given)

			// THEN
			assert.Equal(t, tc.expected, gql)
		})
	}

}

func TestLimitRangeConverter_ToGQLs(t *testing.T) {
	// GIVEN
	conv := limitRangeConverter{}

	// WHEN
	gql := conv.ToGQLs(fixLimitRanges())

	// THEN
	assert.Equal(t, []gqlschema.LimitRange{
		*fixGQLLimitRange(),
	}, gql)
}

func fixLimitRange() *v1.LimitRange {
	return &v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fixLimitRangeName(),
			Namespace: fixLimitRangeNamespace(),
		},
		Spec: v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{
				{
					Max: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse(*fixLimits()[0].Max.Memory),
						v1.ResourceCPU:    resource.MustParse(*fixLimits()[0].Max.Cpu),
					},
					Default: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse(*fixLimits()[0].Default.Memory),
						v1.ResourceCPU:    resource.MustParse(*fixLimits()[0].Default.Cpu),
					},
					DefaultRequest: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse(*fixLimits()[0].DefaultRequest.Memory),
						v1.ResourceCPU:    resource.MustParse(*fixLimits()[0].DefaultRequest.Cpu),
					},
				},
			},
		},
	}
}

func fixLimitRanges() []*v1.LimitRange {
	return []*v1.LimitRange{
		fixLimitRange(),
	}
}

func fixGQLLimitRange() *gqlschema.LimitRange {
	return &gqlschema.LimitRange{
		Name:   fixLimitRangeName(),
		Limits: fixLimits(),
	}
}

func fixLimitRangeName() string {
	return "kyma-default"
}

func fixLimitRangeNamespace() string {
	return "kyma-integration"
}

func fixLimits() []gqlschema.LimitRangeItem {
	return []gqlschema.LimitRangeItem{
		{
			Max: gqlschema.ResourceType{
				Memory: ptrStr("120Mi"),
				Cpu:    ptrStr("100m"),
			},
			Default: gqlschema.ResourceType{
				Memory: ptrStr("120Mi"),
				Cpu:    ptrStr("10"),
			},
			DefaultRequest: gqlschema.ResourceType{
				Memory: ptrStr("120Mi"),
				Cpu:    ptrStr("1200m"),
			},
		},
	}
}

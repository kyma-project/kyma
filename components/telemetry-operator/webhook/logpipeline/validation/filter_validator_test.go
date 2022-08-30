package validation

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateCustomFilter(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "foo"},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    Name    http`,
			},
		},
	}

	sut := NewFilterValidator()
	err := sut.Validate(logPipeline)
	require.NoError(t, err)
}

func TestValidateCustomFiltersContainsNoName(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{Custom: `
    Match   *`,
				},
			},
		},
	}

	sut := NewFilterValidator()
	err := sut.Validate(logPipeline)

	require.Error(t, err)
	require.Contains(t, err.Error(), "configuration section does not have name attribute")
}

func TestValidateCustomFiltersContainsMatch(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{Custom: `
    Name    grep
    Match   *`,
				},
			},
		},
	}

	sut := NewFilterValidator()
	err := sut.Validate(logPipeline)

	require.Error(t, err)
	require.Contains(t, err.Error(), "plugin 'grep' contains match condition. Match conditions are forbidden")
}

func TestDeniedFilterPlugins(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "foo"},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Filters: []telemetryv1alpha1.Filter{
				{Custom: `
    Name    lua`,
				},
			},
		},
	}

	sut := NewFilterValidator("lua", "multiline")
	err := sut.Validate(logPipeline)

	require.Error(t, err)
	require.Contains(t, err.Error(), "plugin 'lua' is forbidden. ")
}

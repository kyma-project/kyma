package validation

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestValidateOutput(t *testing.T) {
	outputValidator := NewOutputValidator()

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    name    http`,
			},
		},
	}

	err := outputValidator.Validate(logPipeline)
	require.NoError(t, err)
}

func TestValidateOutputFail(t *testing.T) {
	outputValidator := NewOutputValidator()

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Custom: `
    name    http
	storage.total_limit_size 10G`,
			},
		},
	}

	err := outputValidator.Validate(logPipeline)
	require.Error(t, err)
}

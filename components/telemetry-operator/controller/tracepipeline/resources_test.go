package tracepipeline

import (
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestMakeConfigMap(t *testing.T) {
	output := v1alpha1.TracePipelineOutput{
		Otlp: v1alpha1.OtlpOutput{
			Endpoint: v1alpha1.ValueType{
				Value: "localhost",
			},
		},
	}
	cm := makeConfigMap(output)
	require.NotNil(t, cm)
}

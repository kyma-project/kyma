package validation

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestValidateParser(t *testing.T) {
	parserValidator := NewParserValidator()

	logParser := &telemetryv1alpha1.LogParser{
		Spec: telemetryv1alpha1.LogParserSpec{
			Parser: `
      Format            json
      Time_Key          time
      Time_Format %Y-%m-%dT%H:%M:%S`,
		},
	}

	err := parserValidator.Validate(logParser)
	require.NoError(t, err)
}

func TestValidateParserFail(t *testing.T) {
	parserValidator := NewParserValidator()

	logParser := &telemetryv1alpha1.LogParser{
		Spec: telemetryv1alpha1.LogParserSpec{
			Parser: `
      Name              foo
      Format            json
      Time_Key          time
      Time_Format %Y-%m-%dT%H:%M:%S`,
		},
	}

	err := parserValidator.Validate(logParser)
	require.Error(t, err)
}

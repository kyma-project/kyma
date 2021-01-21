package logger_test

import (
	"github.com/kyma-project/kyma/common/logger/logger"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatMapping(t *testing.T) {

	testCases := []struct {
		name        string
		input       string
		expected    logger.Format
		expectedErr bool
	}{
		{
			name:        "text format",
			input:       "text",
			expected:    logger.TEXT,
			expectedErr: false,
		},
		{
			name:        "json format",
			input:       "json",
			expected:    logger.JSON,
			expectedErr: false,
		},
		{
			name:        "not existing format",
			input:       "csv",
			expectedErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//WHEN

			output, err := logger.MapFormat(testCase.input)

			//THEN
			if !testCase.expectedErr {
				assert.Equal(t, testCase.expected, output)
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

		})
	}
}

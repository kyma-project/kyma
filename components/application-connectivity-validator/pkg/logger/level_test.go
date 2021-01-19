package logger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/pkg/logger"
)

func TestLevelMapping(t *testing.T) {

	testCases := []struct {
		name        string
		input       string
		expected    logger.Level
		expectedErr bool
	}{
		{
			name:        "debug level",
			input:       "debug",
			expected:    logger.DEBUG,
			expectedErr: false,
		},
		{
			name:        "info level",
			input:       "info",
			expected:    logger.INFO,
			expectedErr: false,
		},
		{
			name:        "warn level",
			input:       "warn",
			expected:    logger.WARN,
			expectedErr: false,
		},
		{
			name:        "error level",
			input:       "error",
			expected:    logger.ERROR,
			expectedErr: false,
		},
		{
			name:        "not existing level",
			input:       "level",
			expectedErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//WHEN

			output, err := logger.MapLevel(testCase.input)

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

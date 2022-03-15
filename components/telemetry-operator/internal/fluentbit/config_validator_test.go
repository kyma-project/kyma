package fluentbit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractError(t *testing.T) {

	testCases := []struct {
		name          string
		output        string
		expectedError string
	}{
		{
			name:          "No error present",
			output:        "configuration test is successful",
			expectedError: "",
		}, {
			name:          "Single line error present",
			output:        "Error: Invalid flush value. Aborting",
			expectedError: "Invalid flush value.",
		},
		{
			name:          "Multiline error present",
			output:        "Error setting up tail.0 plugin property 'Mem_Buf_Limit'\nError: Configuration file contains errors. Aborting\n",
			expectedError: "Error setting up tail.0 plugin property 'Mem_Buf_Limit'",
		},
		{
			name:          "Multiline error with logs present",
			output:        "[2022/03/14 10:51:59] [  Error] File dynamic-parsers/parsers.conf\n[2022/03/14 10:51:59] [  Error] Error in line 4: Invalid indentation level",
			expectedError: "Error in line 4: Invalid indentation level",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := extractError(tc.output)
			assert.Equal(t, tc.expectedError, err, "invalid error extracted")
		})
	}
}

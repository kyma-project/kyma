package fluentbit

import (
	"os"
	"strings"
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
		},
		{
			name:          "Single line error present",
			output:        "\u001b[1m\u001B[91mError\u001B[0m: Invalid flush value. Aborting",
			expectedError: "Invalid flush value.",
		},
		{
			name:          "Multiline error present",
			output:        "Error setting up tail.0 plugin property 'Mem_Buf_Limit'\nError: Configuration file contains errors. Aborting",
			expectedError: "Error setting up tail.0 plugin property 'Mem_Buf_Limit'",
		},
		{
			name:          "Multiline error with logs present",
			output:        "[2022/03/21 09:36:37] [  Error] File /tmp/dry-rundabf0a0d-f27c-4bb8-860e-e90553aa6ef8/dynamic/logpipeline.conf\n[2022/03/21 09:36:37] [  Error] Error in line 3: Invalid indentation level\nError: Configuration file contains errors. Aborting",
			expectedError: "Error in line 3: Invalid indentation level",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := extractError(tc.output)
			assert.Equal(t, tc.expectedError, err, "invalid error extracted")
		})
	}
}

func TestListPlugins(t *testing.T) {
	plugins := []string{"flb-out_sequentialhttp.so", "out_grafana_loki.so"}
	dir, err := os.MkdirTemp("", "plugins")
	assert.NoError(t, err)

	defer os.RemoveAll(dir)

	for _, plugin := range plugins {
		_, err := os.Create(dir + "/" + plugin)
		assert.NoError(t, err)
	}

	// should be ignored
	err = os.Mkdir(dir+"/test", 0777)
	assert.NoError(t, err)

	actual, err := listPlugins(dir)
	assert.NoError(t, err)
	assert.Equal(t, len(plugins), len(actual))

	for _, found := range actual {
		assert.True(t, strings.HasPrefix(found, dir))
		file := strings.TrimPrefix(found, dir+"/")
		assert.Contains(t, plugins, file)
	}

	_, err = listPlugins("/not/existing/dir")
	assert.Error(t, err)
}

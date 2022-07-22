package validation

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
			name:          "Invalid Flush value",
			output:        "[2022/05/24 16:20:55] [\u001b[1m\u001B[91merror\u001B[0m] invalid flush value, aborting.",
			expectedError: "invalid flush value, aborting.",
		},
		{
			name:          "Plugin does not exist",
			output:        "[2022/05/24 16:54:56] [error] [config] section 'abc' tried to instance a plugin name that don't exists\n[2022/05/24 16:54:56] [error] configuration file contains errors, aborting.",
			expectedError: "section 'abc' tried to instance a plugin name that don't exists",
		},
		{
			name:          "Invalid memory buffer limit",
			output:        "[2022/05/24 15:56:05] [error] [config] could not configure property 'Mem_Buf_Limit' on input plugin with section name 'tail'\nconfiguration test is successful",
			expectedError: "could not configure property 'Mem_Buf_Limit' on input plugin with section name 'tail'",
		},
		{
			name:          "Invalid indentation level",
			output:        "[2022/05/24 15:59:59] [error] [config] error in dynamic-parsers/parsers.conf:3: invalid indentation level\n",
			expectedError: "invalid indentation level",
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

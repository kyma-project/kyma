package renderer

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderer_RenderTemplate(t *testing.T) {
	// GIVEN
	renderer := New("testdata")

	testString := "test"

	tests := []struct {
		name         string
		templateName TemplateName
		testData     interface{}
		output       string
		err          error
	}{
		{
			name:         "success",
			templateName: "a",
			testData:     nil,
			output:       "HELLO",
			err:          nil,
		},
		{
			name:         "success with variable",
			templateName: "b",
			testData: struct {
				Test string
			}{
				Test: testString,
			},
			output: testString,
			err:    nil,
		},
		{
			name:         "error when invalid variable",
			templateName: "b",
			testData:     "error",
			output:       "",
			err:          errors.New(`executing "b" at <.Test>: can't evaluate field Test in type string`),
		},
	}
	for _, tt := range tests {
		out := &bytes.Buffer{}

		// WHEN
		err := renderer.RenderTemplate(out, tt.templateName, tt.testData)

		// THEN
		if tt.err == nil {
			require.NoError(t, err)
			assert.Equal(t, tt.output, out.String())
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), tt.err.Error())
			assert.Empty(t, out)
		}
	}
}

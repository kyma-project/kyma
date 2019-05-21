package gqlschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplicationMappingService_UnmarshalGQL(t *testing.T) {
	// GIVEN
	for name, tc := range map[string]struct {
		input    interface{}
		err      bool
		errMsg   string
		expected ApplicationMappingService
	}{
		"no error": {
			input:    map[string]interface{}{"id": "1234-abcd"},
			err:      false,
			expected: ApplicationMappingService{ID: "1234-abcd"},
		},
		"input error": {
			input:  "wrong input",
			err:    true,
			errMsg: "unexpected services type: string, should be map[string]string",
		},
		"missing parameter error": {
			input:  map[string]interface{}{"wrong_param": "1234-abcd"},
			err:    true,
			errMsg: "missing services id parameter",
		},
		"wrong parameter value": {
			input:  map[string]interface{}{"id": 1234},
			err:    true,
			errMsg: "unexpected services type: int, should be string",
		},
	} {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ams := ApplicationMappingService{}

			// WHEN
			err := ams.UnmarshalGQL(tc.input)

			// THEN
			if tc.err {
				assert.EqualError(t, err, tc.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, ams)
			}
		})
	}
}

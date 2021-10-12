package normalization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeServiceName(t *testing.T) {

	tests := []struct {
		name           string
		displayName    string
		id             string
		expected       string
		expectedWithId string
	}{
		{
			name:           "should replace upper case with lower case",
			displayName:    "DisplayName",
			id:             "id",
			expected:       "displayname",
			expectedWithId: "displayname-87ea5",
		},
		{
			name:           "should replace non alpha numeric characters with --",
			displayName:    "display_!@#$%^&*()name",
			id:             "id",
			expected:       "display-name",
			expectedWithId: "display-name-87ea5",
		},
		{
			name:           "should remove leading dashes",
			displayName:    "-----displayname",
			id:             "id",
			expected:       "displayname",
			expectedWithId: "displayname-87ea5",
		},
		{
			name:           "should trim if name too long",
			displayName:    "VeryVeryVeryVeryVeryVeryVEryVeryVeryVeryVeryVeryVeryVeryLongDescription",
			id:             "id",
			expected:       "veryveryveryveryveryveryveryveryveryveryveryveryveryveryl",
			expectedWithId: "veryveryveryveryveryveryveryveryveryveryveryveryveryveryl-87ea5",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			normalised := NormalizeName(tc.displayName)
			assert.Equal(t, tc.expected, normalised)

			normalisedWithId := NormalizeServiceNameWithId(tc.displayName, tc.id)
			assert.Equal(t, tc.expectedWithId, normalisedWithId)
		})
	}
}

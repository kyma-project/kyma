package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRequestParametersIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		rp    RequestParameters
		empty bool
	}{
		{name: "nil values", rp: RequestParameters{}, empty: true},
		{name: "empty values", rp: RequestParameters{Headers: &map[string][]string{}, QueryParameters: &map[string][]string{}}, empty: true},
		{name: "has header", rp: RequestParameters{Headers: &map[string][]string{"header": {}}, QueryParameters: &map[string][]string{}}, empty: false},
		{name: "has param", rp: RequestParameters{Headers: &map[string][]string{}, QueryParameters: &map[string][]string{"param": {}}}, empty: false},
		{name: "has header and param", rp: RequestParameters{Headers: &map[string][]string{"header": {}}, QueryParameters: &map[string][]string{"param": {}}}, empty: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			a.Equal(tc.empty, tc.rp.IsEmpty())
		})
	}
}

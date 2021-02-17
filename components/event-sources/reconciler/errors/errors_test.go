package errors

import (
	"context"
	"errors"
	"testing"
)

func TestHandle(t *testing.T) {
	dummyErr := errors.New("error")

	testCases := map[string]struct {
		input, expect error
	}{
		"Nil input": {
			nil, nil,
		},
		"Skippable error gets ignored": {
			NewSkippable(dummyErr), nil,
		},
		"Non-skippable error gets returned": {
			dummyErr, dummyErr,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(*testing.T) {
			ctx := context.Background()
			if expect, got := tc.expect, Handle(tc.input, ctx, ""); got != expect {
				t.Errorf("Expected %v, got %v", expect, got)
			}
		})
	}
}

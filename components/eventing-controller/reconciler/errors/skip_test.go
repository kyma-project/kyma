package errors

import (
	"errors"
	"fmt"
	"testing"
)

func Test_NewSkippable(t *testing.T) {
	testCases := []struct {
		error error
	}{
		{error: NewSkippable(nil)},
		{error: NewSkippable(NewSkippable(nil))},
		{error: NewSkippable(fmt.Errorf("some error"))},
		{error: NewSkippable(NewSkippable(fmt.Errorf("some error")))},
	}

	for _, tc := range testCases {
		skippableErr := NewSkippable(tc.error)
		if skippableErr == nil {
			t.Errorf("Test NewSkippable retuned nil error")
			continue
		}
		if err := errors.Unwrap(skippableErr); tc.error != err {
			t.Errorf("Test NewSkippable failed, want: %#v but got: %#v", tc.error, err)
		}
	}
}

func Test_IsSkippable(t *testing.T) {
	testCases := []struct {
		name          string
		givenError    error
		wantSkippable bool
	}{
		{
			name:          "nil error, should be skipped",
			givenError:    nil,
			wantSkippable: true,
		},
		{
			name:          "skippable error, should be skipped",
			givenError:    NewSkippable(fmt.Errorf("some errore")),
			wantSkippable: true,
		},
		{
			name:          "not-skippable error, should not be skipped",
			givenError:    fmt.Errorf("some error"),
			wantSkippable: false,
		},
		{
			name:          "not-skippable error which wraps a skippable error, should not be skipped",
			givenError:    fmt.Errorf("some error %w", NewSkippable(fmt.Errorf("some error"))),
			wantSkippable: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if gotSkippable := IsSkippable(tc.givenError); tc.wantSkippable != gotSkippable {
				t.Errorf("Test skippable failed, want: %v but got: %v", tc.wantSkippable, gotSkippable)
			}
		})
	}
}

//go:build unit
// +build unit

package errors_test

import (
	"errors"
	"fmt"
	"testing"

	ctrlserrors "github.com/kyma-project/kyma/components/eventing-controller/controllers/errors"
)

func Test_NewSkippable(t *testing.T) {
	testCases := []struct {
		error error
	}{
		{error: ctrlserrors.NewSkippable(nil)},
		{error: ctrlserrors.NewSkippable(ctrlserrors.NewSkippable(nil))},
		{error: ctrlserrors.NewSkippable(fmt.Errorf("some error"))},
		{error: ctrlserrors.NewSkippable(ctrlserrors.NewSkippable(fmt.Errorf("some error")))},
	}

	for _, tc := range testCases {
		skippableErr := ctrlserrors.NewSkippable(tc.error)

		if skippableErr == nil {
			t.Errorf("test NewSkippable retuned nil error")
			continue
		}
		if err := errors.Unwrap(skippableErr); tc.error != err {
			t.Errorf("test NewSkippable failed, want: %#v but got: %#v", tc.error, err)
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
			givenError:    ctrlserrors.NewSkippable(fmt.Errorf("some errore")),
			wantSkippable: true,
		},
		{
			name:          "not-skippable error, should not be skipped",
			givenError:    fmt.Errorf("some error"),
			wantSkippable: false,
		},
		{
			name:          "not-skippable error which wraps a skippable error, should not be skipped",
			givenError:    fmt.Errorf("some error %w", ctrlserrors.NewSkippable(fmt.Errorf("some error"))),
			wantSkippable: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if gotSkippable := ctrlserrors.IsSkippable(tc.givenError); tc.wantSkippable != gotSkippable {
				t.Errorf("test skippable failed, want: %v but got: %v", tc.wantSkippable, gotSkippable)
			}
		})
	}
}

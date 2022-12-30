package errors_test

import (
	"errors"
	"fmt"
	"testing"

	pkgerrors "github.com/kyma-project/kyma/components/eventing-controller/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var errInvalidStorageType = pkgerrors.NewArgumentError("invalid stream storage type: %q")

func ExampleMakeError() {
	actualError := errors.New("failed to connect to NATS")
	underlyingError := errors.New("some error coming from the NATS package")
	fmt.Println(pkgerrors.MakeError(actualError, underlyingError))
	// Output: failed to connect to NATS: some error coming from the NATS package
}

func ExampleArgumentError() {
	e := errInvalidStorageType
	e = e.WithArg("some storage type")

	fmt.Println(e.Error())
	// Output: invalid stream storage type: "some storage type"
}

func Test_ArgumentError_Is(t *testing.T) {
	t.Parallel()

	const givenStorageType = "some storage type"

	testCases := []struct {
		name       string
		givenError func() error
		wantIsTrue bool
	}{
		{
			name: "with argument",
			givenError: func() error {
				return errInvalidStorageType.WithArg(givenStorageType)
			},
			wantIsTrue: true,
		},
		{
			name: "with argument and wrapped",
			givenError: func() error {
				e := errInvalidStorageType.WithArg(givenStorageType)
				return fmt.Errorf("%v: %w", errors.New("new error"), e)
			},
			wantIsTrue: true,
		},
		{
			name: "without argument",
			givenError: func() error {
				return errInvalidStorageType
			},
			wantIsTrue: true,
		},
		{
			name: "completely different error",
			givenError: func() error {
				return errors.New("unknown error")
			},
			wantIsTrue: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// when
			ok := errors.Is(tc.givenError(), errInvalidStorageType)

			// then
			assert.Equal(t, ok, tc.wantIsTrue)
		})
	}
}

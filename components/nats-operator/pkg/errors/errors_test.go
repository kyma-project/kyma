package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverable(t *testing.T) {
	// given
	err := Recoverable(fmt.Errorf("error"))

	// when
	err1, ok1 := err.(error)
	err2, ok2 := err.(recoverable)

	// expect
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.NotNil(t, err1)
	assert.NotNil(t, err2)
	assert.NotEmpty(t, err1.Error())
	assert.NotEmpty(t, err2.Error())
}

func TestIsRecoverable(t *testing.T) {
	testCases := []struct {
		name            string
		error           error
		wantRecoverable bool
	}{
		{
			name:            "non-recoverable error",
			error:           fmt.Errorf("error"),
			wantRecoverable: false,
		},
		{
			name:            "recoverable error",
			error:           Recoverable(fmt.Errorf("error")),
			wantRecoverable: true,
		},
		{
			name:            "recoverable nil error",
			error:           Recoverable(nil),
			wantRecoverable: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			gotRecoverable := IsRecoverable(tc.error)

			// expect
			assert.Equal(t, tc.wantRecoverable, gotRecoverable)
		})
	}
}

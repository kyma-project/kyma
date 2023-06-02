//go:build unit

package jetstream

import (
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/stretchr/testify/assert"
)

func TestUnitValidate_For_Errors(t *testing.T) {
	tests := []struct {
		name        string
		givenConfig env.NATSConfig
		wantError   error
	}{
		{
			name:        "ErrorEmptyStream",
			givenConfig: env.NATSConfig{JSStreamName: ""},
			wantError:   ErrEmptyStreamName,
		},
		{
			name:        "ErrorStreamToLong",
			givenConfig: env.NATSConfig{JSStreamName: fixtureStreamNameTooLong()},
			wantError:   ErrStreamNameTooLong,
		},
		{
			name: "ErrorStorageType",
			givenConfig: env.NATSConfig{
				JSStreamName:        "not-empty",
				JSStreamStorageType: "invalid-storage-type",
			},
			wantError: ErrInvalidStorageType.WithArg("invalid-storage-type"),
		},
		{
			name: "ErrorRetentionPolicy",
			givenConfig: env.NATSConfig{
				JSStreamName:            "not-empty",
				JSStreamStorageType:     StorageTypeMemory,
				JSStreamRetentionPolicy: "invalid-retention-policy",
			},
			wantError: ErrInvalidRetentionPolicy.WithArg("invalid-retention-policy"),
		},
		{
			name: "ErrorDiscardPolicy",
			givenConfig: env.NATSConfig{
				JSStreamName:            "not-empty",
				JSStreamStorageType:     StorageTypeMemory,
				JSStreamRetentionPolicy: RetentionPolicyInterest,
				JSStreamDiscardPolicy:   "invalid-discard-policy",
			},
			wantError: ErrInvalidDiscardPolicy.WithArg("invalid-discard-policy"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.givenConfig)
			assert.ErrorIs(t, err, tc.wantError)
		})
	}
}

func fixtureStreamNameTooLong() string {
	b := strings.Builder{}
	for i := 0; i < (jsMaxStreamNameLength + 1); i++ {
		b.WriteString("a")
	}
	streamName := b.String()
	return streamName
}

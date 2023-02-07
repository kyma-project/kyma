package jetstream

import (
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func Test_streamIsConfiguredCorrectly(t *testing.T) {
	streamConfig := &nats.StreamConfig{
		Name:      "test",
		Storage:   nats.FileStorage,
		Replicas:  3,
		Retention: nats.InterestPolicy,
		MaxMsgs:   10,
		MaxBytes:  1024,
		Discard:   nats.DiscardNew,
		Subjects:  []string{"kyma.>"},
	}

	testCases := []struct {
		name            string
		ecDefinedConfig nats.StreamConfig
		natsConfig      nats.StreamConfig
		wantResult      bool
	}{
		{
			name:            "same configs should return true",
			ecDefinedConfig: *streamConfig,
			natsConfig:      *streamConfig,
			wantResult:      true,
		},
		{
			name:            "Non relevant nats config should not effect result",
			ecDefinedConfig: *streamConfig,
			natsConfig: nats.StreamConfig{
				Name:              streamConfig.Name,
				Storage:           streamConfig.Storage,
				Replicas:          streamConfig.Replicas,
				Retention:         streamConfig.Retention,
				MaxMsgs:           streamConfig.MaxMsgs,
				MaxBytes:          streamConfig.MaxBytes,
				Discard:           streamConfig.Discard,
				Subjects:          streamConfig.Subjects,
				MaxMsgsPerSubject: 99,
			},
			wantResult: true,
		},
		{
			name:            "Different nats config should return false",
			ecDefinedConfig: *streamConfig,
			natsConfig: nats.StreamConfig{
				Name:      "test",
				Storage:   nats.FileStorage,
				Replicas:  3,
				Retention: nats.InterestPolicy,
				MaxMsgs:   10,
				MaxBytes:  2048,
				Discard:   nats.DiscardNew,
				Subjects:  []string{"kyma.>"},
			},
			wantResult: false,
		},
		{
			name:            "Different subject config should return false",
			ecDefinedConfig: *streamConfig,
			natsConfig: nats.StreamConfig{
				Name:      "test",
				Storage:   nats.FileStorage,
				Replicas:  3,
				Retention: nats.InterestPolicy,
				MaxMsgs:   10,
				MaxBytes:  2048,
				Discard:   nats.DiscardNew,
				Subjects:  []string{"xyz.>"},
			},
			wantResult: false,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantResult, streamIsConfiguredCorrectly(tc.natsConfig, tc.ecDefinedConfig))
		})
	}
}

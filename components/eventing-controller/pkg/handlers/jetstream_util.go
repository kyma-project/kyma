package handlers

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

const (
	JetStreamStorageTypeMemory = "memory"
	JetStreamStorageTypeFile   = "file"

	JetStreamRetentionPolicyLimits   = "limits"
	JetStreamRetentionPolicyInterest = "interest"
)

func toJetStreamStorageType(s string) (nats.StorageType, error) {
	switch s {
	case JetStreamStorageTypeMemory:
		return nats.MemoryStorage, nil
	case JetStreamStorageTypeFile:
		return nats.FileStorage, nil
	}
	return nats.MemoryStorage, fmt.Errorf("invalid stream storage type %q", s)
}

func toJetStreamRetentionPolicy(s string) (nats.RetentionPolicy, error) {
	switch s {
	case JetStreamRetentionPolicyLimits:
		return nats.LimitsPolicy, nil
	case JetStreamRetentionPolicyInterest:
		return nats.InterestPolicy, nil
	}
	return nats.LimitsPolicy, fmt.Errorf("invalid stream retention policy %q", s)
}

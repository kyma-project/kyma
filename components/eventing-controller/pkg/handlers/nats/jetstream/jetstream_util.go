package jetstream

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

const (
	StorageTypeMemory = "memory"
	StorageTypeFile   = "file"

	RetentionPolicyLimits   = "limits"
	RetentionPolicyInterest = "interest"

	ConsumerDeliverPolicyAll            = "all"
	ConsumerDeliverPolicyLast           = "last"
	ConsumerDeliverPolicyLastPerSubject = "last_per_subject"
	ConsumerDeliverPolicyNew            = "new"
)

func toJetStreamStorageType(s string) (nats.StorageType, error) {
	switch s {
	case StorageTypeMemory:
		return nats.MemoryStorage, nil
	case StorageTypeFile:
		return nats.FileStorage, nil
	}
	return nats.MemoryStorage, fmt.Errorf("invalid stream storage type %q", s)
}

func toJetStreamRetentionPolicy(s string) (nats.RetentionPolicy, error) {
	switch s {
	case RetentionPolicyLimits:
		return nats.LimitsPolicy, nil
	case RetentionPolicyInterest:
		return nats.InterestPolicy, nil
	}
	return nats.LimitsPolicy, fmt.Errorf("invalid stream retention policy %q", s)
}

// toJetStreamConsumerDeliverPolicyOpt returns a nats.DeliverPolicy opt based on the given deliver policy string value.
// It returns "DeliverNew" as the default nats.DeliverPolicy opt, if the given deliver policy value is not supported.
// Supported deliver policy values are ("all", "last", "last_per_subject" and "new").
func toJetStreamConsumerDeliverPolicyOptOrDefault(deliverPolicy string) nats.SubOpt {
	switch deliverPolicy {
	case ConsumerDeliverPolicyAll:
		return nats.DeliverAll()
	case ConsumerDeliverPolicyLast:
		return nats.DeliverLast()
	case ConsumerDeliverPolicyLastPerSubject:
		return nats.DeliverLastPerSubject()
	case ConsumerDeliverPolicyNew:
		return nats.DeliverNew()
	}
	return nats.DeliverNew()
}

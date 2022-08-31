package jetstream

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

const (
	JetStreamStorageTypeMemory = "memory"
	JetStreamStorageTypeFile   = "file"

	JetStreamRetentionPolicyLimits   = "limits"
	JetStreamRetentionPolicyInterest = "interest"

	JetStreamConsumerDeliverPolicyAll            = "all"
	JetStreamConsumerDeliverPolicyLast           = "last"
	JetStreamConsumerDeliverPolicyLastPerSubject = "last_per_subject"
	JetStreamConsumerDeliverPolicyNew            = "new"
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

// toJetStreamConsumerDeliverPolicyOpt returns a nats.DeliverPolicy opt based on the given deliver policy string value.
// It returns "DeliverNew" as the default nats.DeliverPolicy opt, if the given deliver policy value is not supported.
// Supported deliver policy values are ("all", "last", "last_per_subject" and "new").
func toJetStreamConsumerDeliverPolicyOptOrDefault(deliverPolicy string) nats.SubOpt {
	switch deliverPolicy {
	case JetStreamConsumerDeliverPolicyAll:
		return nats.DeliverAll()
	case JetStreamConsumerDeliverPolicyLast:
		return nats.DeliverLast()
	case JetStreamConsumerDeliverPolicyLastPerSubject:
		return nats.DeliverLastPerSubject()
	case JetStreamConsumerDeliverPolicyNew:
		return nats.DeliverNew()
	}
	return nats.DeliverNew()
}

package jetstreamv2

// maxJetStreamConsumerNameLength is the maximum preferred length for the JetStream consumer names
// as per https://docs.nats.io/running-a-nats-service/nats_admin/jetstream_admin/naming
const (
	subName                        = "subName"
	subNamespace                   = "subNamespace"
	maxJetStreamConsumerNameLength = 32
)

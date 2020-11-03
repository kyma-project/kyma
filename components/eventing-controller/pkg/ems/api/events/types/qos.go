package types

type Qos string

const (
	QosAtMostOnce  Qos = "AT_MOST_ONCE"
	QosAtLeastOnce Qos = "AT_LEAST_ONCE"
)

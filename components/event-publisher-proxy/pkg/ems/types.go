package ems

type Qos string

const (
	// QosAtLeastOnce the quality of service supported by EMS to send Events with at least once guarantee.
	QosAtLeastOnce Qos = "AT_LEAST_ONCE"
)

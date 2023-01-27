package types

import "strings"

type Qos string

const (
	QosAtMostOnce  Qos = "AT_MOST_ONCE"
	QosAtLeastOnce Qos = "AT_LEAST_ONCE"
)

func IsInvalidQoS(value string) bool {
	value = strings.ReplaceAll(value, "-", "_")
	switch value {
	case string(QosAtLeastOnce), string(QosAtMostOnce):
		return false
	default:
		return true
	}
}

func GetQos(qosStr string) Qos {
	qosStr = strings.ReplaceAll(qosStr, "-", "_")
	switch qosStr {
	case string(QosAtMostOnce):
		return QosAtMostOnce
	default:
		return QosAtLeastOnce
	}
}

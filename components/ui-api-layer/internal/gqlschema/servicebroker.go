package gqlschema

import "time"

type ServiceBroker struct {
	Name              string
	CreationTimestamp time.Time
	Url               string
	Labels            JSON
	Status            ServiceBrokerStatus
}

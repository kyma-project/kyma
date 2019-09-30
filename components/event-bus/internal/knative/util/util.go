package util

import (
	"fmt"
	"strings"
)

const (
	delimiter               = "--"
	defaultChannelNamespace = "kyma-system"
	// SubscriptionSourceID is the key for source id in label
	SubscriptionSourceID = "kyma-source-id"
	// SubscriptionEventType is the key for event type in label
	SubscriptionEventType = "kyma-event-type"
	// SubscriptionEventTypeVersion is the key for event type version in label
	SubscriptionEventTypeVersion = "kyma-event-type-version"
	// SubNs is the key for namespace of the subscription
	SubNs = "kyma-ns"
)

var (
	replacer = strings.NewReplacer("-", "-d", ".", "-p")
)

// GetKnSubscriptionName joins the kySubscriptionName and kySubscriptionNamespace
func GetKnSubscriptionName(kySubscriptionName, kySubscriptionNamespace *string) string {
	return fmt.Sprintf("%s%s%s", *kySubscriptionName, delimiter, escapeHyphensAndPeriods(kySubscriptionNamespace))
}

// GetDefaultChannelNamespace returns the default namespace of Knative/Eventing channels and subscriptions
func GetDefaultChannelNamespace() string {
	return defaultChannelNamespace
}

// The escapeHyphensAndPeriods function applies the following rules in order:
//  * In case there was a '-' or more in any of the argument values, each occurrence of the '-' will be escaped by '-d'.
//  * In case there was a '.' or more in any of the argument values, each occurrence of the '.' will be replaced by the '-p' character sequence,
//    because of a limitation in the current knative version, if the channel name has a '.', the corresponding istio-virtualservice will not be created.
func escapeHyphensAndPeriods(str *string) string {
	return replacer.Replace(*str)
}

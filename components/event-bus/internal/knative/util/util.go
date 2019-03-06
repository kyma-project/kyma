package util

import (
	"fmt"
	"strings"
)

var (
	replacer = strings.NewReplacer("-", "--", ".", "-dot-")
)

// The function applies the following rules in order:
//  * In case there was a '-' or more in any of the argument values, each occurrence of the '-' will be escaped by '--'.
//  * In case there was a '.' or more in any of the argument values, each occurrence of the '.' will be replaced by the '-dot-' character sequence, because of a limitation
//    in the current knative version, if the channel name has a '.', the corresponding istio-virtualservice will not be created.
func escapeHyphensAndPeriods(str *string) string {
	return replacer.Replace(*str)
}

// GetChannelName function joins the sourceID, eventType and eventTypeVersion respectively with a '-' as a delimiter.
func GetChannelName(sourceID, eventType, eventTypeVersion *string) string {
	return fmt.Sprintf("%s-%s-%s", escapeHyphensAndPeriods(sourceID), escapeHyphensAndPeriods(eventType), escapeHyphensAndPeriods(eventTypeVersion))
}

// GetSubscriptionName joins the kySubscriptionName and kySubscriptionNamespace
func GetSubscriptionName(kySubscriptionName, kySubscriptionNamespace *string) string {
	return fmt.Sprintf("%s-%s", *kySubscriptionName, escapeHyphensAndPeriods(kySubscriptionNamespace))
}

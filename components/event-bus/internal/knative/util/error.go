package util

import (
	"github.com/kyma-project/kyma/components/event-bus/api/publish"
)

const (
	// FieldKnativeChannelName knative channel name field
	FieldKnativeChannelName = "knative-channel-name"
)

// ErrorInvalidChannelNameLength returns a publish.Error instance.
func ErrorInvalidChannelNameLength(channelNameMaxLength int) *publish.Error {
	return publish.ErrorInvalidFieldLength(FieldKnativeChannelName, channelNameMaxLength)
}

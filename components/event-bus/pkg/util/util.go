package util

import (
	"fmt"

	"github.com/kyma-project/kyma/components/event-bus/internal/knative/hash"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

// GetChannelName function returns a unique hash for the knative channel name from the given event components
// because of a limitation in the current knative version, if the channel name starts with a number, the corresponding
// istio-virtualservice will not be created, in order to mitigate that, we prefix the channel name with the letter 'k'
func GetChannelName(sourceID, eventType, eventTypeVersion *string) string {
	channelName := util.EncodeChannelName(sourceID, eventType, eventTypeVersion)
	return fmt.Sprintf("k%s", hash.ComputeHash(&channelName))
}

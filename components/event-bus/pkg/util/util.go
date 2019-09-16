package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/event-bus/internal/knative/hash"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

var (
	replacer = strings.NewReplacer(".", "-")
)

// GetChannelName function returns a unique hash for the knative channel name from the given event components
// because of a limitation in the current knative version, if the channel name starts with a number, the corresponding
// istio-virtualservice will not be created, in order to mitigate that, we prefix the channel name with the letter 'k'
func GetChannelName(sourceID, eventType, eventTypeVersion *string) string {
	channelName := util.EncodeChannelName(sourceID, eventType, eventTypeVersion)
	return fmt.Sprintf("k%s", hash.ComputeHash(&channelName))[:25]
}

// GetKnativeChannelName function is the new way to create knative channel name
// Existing Implementation to generate channel name creates overwhelming length of string,
// which eventually creates problem in creating gcp-pub/sub service, Hence restricting the channel length
// to 15 characters and due to a timestamp substring in the channel name we are ensured that the channel name
// is unique
func GetKnativeChannelName(sourceID, eventType *string) string {
	chanName := fmt.Sprintf("%s-%s-%s", strconv.FormatInt(time.Now().Unix(), 10),
		replacer.Replace(*sourceID), replacer.Replace(*eventType))[:25]
	return fmt.Sprintf("k%s", chanName)
}

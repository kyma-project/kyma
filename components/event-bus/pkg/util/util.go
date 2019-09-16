package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	replacer = strings.NewReplacer(".", "-")
)

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

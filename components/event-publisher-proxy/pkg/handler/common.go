package handler

import "strings"

const (
	PublishEndpoint          = "/publish"
	LegacyEndpointSuffix     = "/v1/events"
	SubscribedEndpointSuffix = "/v1/events/subscribed"
)

func IsARequestWithCE(uri string) bool {
	return uri == PublishEndpoint
}
func IsARequestWithLegacyEvent(uri string) bool {
	// Assuming the path should be of the form /:application/v1/events
	uriPathSegments := make([]string, 0)

	for _, segment := range strings.Split(uri, "/") {
		if strings.TrimSpace(segment) != "" {
			uriPathSegments = append(uriPathSegments, segment)
		}
	}
	if len(uriPathSegments) != 3 {
		return false
	}
	if !strings.HasSuffix(uri, LegacyEndpointSuffix) {
		return false
	}
	return true
}

func IsARequestForSubscriptions(uri string) bool {
	// Assuming the path should be of the form /:application/v1/events/subscribed
	if !strings.HasSuffix(uri, SubscribedEndpointSuffix) {
		return false
	}
	return true
}

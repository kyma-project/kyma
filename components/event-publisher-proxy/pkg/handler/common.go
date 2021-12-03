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
	return strings.HasSuffix(uri, LegacyEndpointSuffix)
}

func IsARequestForSubscriptions(uri string) bool {
	// Assuming the path should be of the form /:application/v1/events/subscribed
	return strings.HasSuffix(uri, SubscribedEndpointSuffix)
}

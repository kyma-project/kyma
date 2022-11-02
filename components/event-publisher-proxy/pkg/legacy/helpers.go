package legacy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	kymalogger "github.com/kyma-project/kyma/components/eventing-controller/logger"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/api"
)

const (
	// eventTypeFormat is driven by BEB specification.
	// An EventType must have at least 4 segments separated by dots in the form of:
	// <domainNamespace>.<businessObjectName>.<operation>.<version>
	eventTypeFormat = "%s.%s.%s.%s"

	// eventTypeFormatWithoutPrefix must have at least 3 segments separated by dots in the form of:
	// <businessObjectName>.<operation>.<version>
	eventTypeFormatWithoutPrefix = "%s.%s.%s"

	legacyEventsName = "legacy-events"
)

// ParseApplicationNameFromPath returns application name from the URL.
// The format of the URL is: /:application-name/v1/...
func ParseApplicationNameFromPath(path string) string {
	// Assumption: Clients(application validator which has a flag for the path (https://github.com/kyma-project/kyma/blob/main/components/application-connectivity-validator/cmd/applicationconnectivityvalidator/applicationconnectivityvalidator.go#L49) using this endpoint must be sending request to path /:application/v1/events
	// Hence it should be safe to return 0th index as the application name
	pathSegments := make([]string, 0)
	for _, segment := range strings.Split(path, "/") {
		if strings.TrimSpace(segment) != "" {
			pathSegments = append(pathSegments, segment)
		}
	}
	return pathSegments[0]
}

// is2XXStatusCode checks whether status code is a 2XX status code
func is2XXStatusCode(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, resp *api.PublishEventResponses) {
	encoder := json.NewEncoder(w)
	w.Header().Set(internal.HeaderContentType, internal.ContentTypeApplicationJSON)

	if resp.Error != nil {
		w.WriteHeader(resp.Error.Status)
		_ = encoder.Encode(resp.Error)
		return
	}

	if resp.Ok != nil {
		_ = encoder.Encode(resp.Ok)
		return
	}

	// init the contexted logger
	logger, _ := kymalogger.New("json", "error")
	namedLogger := logger.WithContext().Named(legacyEventsName)

	namedLogger.Error("Received an empty response")
}

// formatEventType joins the given prefix, application, eventType and version with a "." as a separator.
// It ignores the prefix if it is empty.
func formatEventType(prefix, application, eventType, version string) string {
	if len(strings.TrimSpace(prefix)) == 0 {
		return fmt.Sprintf(eventTypeFormatWithoutPrefix, application, eventType, version)
	}
	return fmt.Sprintf(eventTypeFormat, prefix, application, eventType, version)
}

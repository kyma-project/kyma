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
	// `<domainNamespace>.<businessObjectName>.<operation>.<version>`.
	eventTypeFormat = "%s.%s.%s.%s"

	// eventTypeFormatWithoutPrefix must have at least 3 segments separated by dots in the form of:
	// `<businessObjectName>.<operation>.<version>`.
	eventTypeFormatWithoutPrefix = "%s.%s.%s"

	legacyEventsName = "legacy-events"
)

// ParseApplicationNameFromPath returns application name from the URL.
// The format of the URL is: /:application-name/v1/...
// returns empty string if application-name cannot be found.
func ParseApplicationNameFromPath(path string) string {
	path = strings.TrimLeft(path, "/")
	application, _, ok := strings.Cut(path, "/")
	if ok {
		return application
	}
	return ""
}

// is2XXStatusCode checks whether status code is a 2XX status code.
func is2XXStatusCode(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

// WriteJSONResponse writes a JSON response.
func WriteJSONResponse(w http.ResponseWriter, resp *api.PublishEventResponses) {
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

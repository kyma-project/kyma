package httptools

import (
	"net/http"

	"go.uber.org/zap"
)

// SetHeaders modifies request headers setting additional entries from customHeaders
func SetHeaders(reqHeaders http.Header, customHeaders *map[string][]string) {
	if customHeaders == nil {
		return
	}

	for header, values := range *customHeaders {
		if _, ok := reqHeaders[header]; ok {
			// if header is already specified we do not interfere with it
			continue
		}

		reqHeaders[header] = values
	}
}

// RemoveHeader modifies request headers removing headerToRemove entry
func RemoveHeader(reqHeaders http.Header, headerToRemove string) {
	if _, ok := reqHeaders[headerToRemove]; ok {
		zap.L().Debug("Removing header",
			zap.String("header", headerToRemove))
		reqHeaders.Del(headerToRemove)
	}
}

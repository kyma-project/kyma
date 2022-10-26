package legacytest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func InvalidLegacyRequest(version, appname, eventType string) (*http.Request, error) {
	body, err := json.Marshal(map[string]string{
		"eventtype":        eventType,
		"eventtypeversion": version,
		"eventtime":        "2020-04-02T21:37:00Z",
		"data":             "{\"legacy\":\"event\"}",
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://localhost:8080/%s/%s/events", appname, version)
	return http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
}

func ValidLegacyRequest(version, appname, eventType string) (*http.Request, error) {
	body, err := json.Marshal(map[string]string{
		"event-type":         eventType,
		"event-type-version": version,
		"event-time":         "2020-04-02T21:37:00Z",
		"data":               "{\"legacy\":\"event\"}",
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://localhost:8080/%s/%s/events", appname, version)
	return http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
}

func InvalidLegacyRequestOrDie(t *testing.T, version, appname, eventType string) *http.Request {
	r, err := InvalidLegacyRequest(version, appname, eventType)
	assert.NoError(t, err)
	return r
}

func ValidLegacyRequestOrDie(t *testing.T, version, appname, eventType string) *http.Request {
	r, err := ValidLegacyRequest(version, appname, eventType)
	assert.NoError(t, err)
	return r
}

package testing

import (
	"bytes"
	"net/http"
)

const (
	// binary cloudevent headers
	CeIDHeader          = "CE-ID"
	CeTypeHeader        = "CE-Type"
	CeSourceHeader      = "CE-Source"
	CeSpecVersionHeader = "CE-SpecVersion"

	// cloudevent attributes
	CeID          = "00000"
	CeType        = "someType"
	CeSource      = "someSource"
	CeSpecVersion = "1.0"
)

func SendEvent(endpoint, body string, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}

	if headers != nil {
		for k, v := range headers {
			req.Header[k] = v
		}
	}

	client := http.Client{}
	defer client.CloseIdleConnections()

	return client.Do(req)
}

func GetStructuredMessageHeaders() http.Header {
	return http.Header{"Content-Type": []string{"application/cloudevents+json"}}
}

func GetBinaryMessageHeaders() http.Header {
	headers := make(http.Header)
	headers.Add(CeIDHeader, CeID)
	headers.Add(CeTypeHeader, CeType)
	headers.Add(CeSourceHeader, CeSource)
	headers.Add(CeSpecVersionHeader, CeSpecVersion)
	return headers
}

package events

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"reflect"
	"testing"

	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/ems"
)

func TestWriteRequestWithHeaders(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Add("Content-Type", "application/cloudevents+json")

	message := cehttp.NewMessageFromHttpRequest(req)
	defer func() { _ = message.Finish(nil) }()

	additionalHeaders := http.Header{
		"qos":    []string{string(ems.QosAtLeastOnce)},
		"accept": []string{"application/json"},
		"key1":   []string{"value1", "value2"},
		"key2":   []string{"value3"},
	}
	additionalHeadersCopy := copyHeaders(additionalHeaders)

	ctx := context.Background()
	err := WriteRequestWithHeaders(ctx, message, req, additionalHeaders)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(additionalHeaders, additionalHeadersCopy) {
		t.Fatal("Write request with headers should not change the input HTTP headers")
	}

	for k, v := range additionalHeaders {
		vv, ok := req.Header[textproto.CanonicalMIMEHeaderKey(k)]
		if !ok || !reflect.DeepEqual(v, vv) {
			t.Fatal("The HTTP request should contain the given HTTP headers")
		}
	}
}

func copyHeaders(headers http.Header) http.Header {
	headersCopy := make(http.Header)
	for k, v := range headers {
		headersCopy[k] = v
	}
	return headersCopy
}

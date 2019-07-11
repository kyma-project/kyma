package v1

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	apiv1 "github.com/kyma-project/kyma/components/event-service/internal/events/api/v1"
	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/event-service/internal/httptools"
)

const (
	maxRequestSize = 64 * 1024
)

func TestEventOk(t *testing.T) {
	saved := handleEvent
	defer func() { handleEvent = saved }()

	handleEvent = func(parameters *apiv1.PublishEventParametersV1, response *api.PublishEventResponses,
		traceHeaders *map[string]string, forwardHeaders *map[string][]string) (err error) {
		ok := api.PublishResponse{EventID: "responseEventId"}
		response.Ok = &ok
		return
	}
	s := "{\"event-type\":\"order.created\",\"event-type-version\":\"v1\",\"event-id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"event-time\":\"2012-11-01T22:08:41+00:00\"}"
	req, err := http.NewRequest(http.MethodPost, shared.EventsV1Path, strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler := NewEventsHandler(maxRequestSize)
	handler.ServeHTTP(recorder, req)
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}
	if contentType := recorder.Result().Header.Get("Content-Type"); contentType != httpconsts.ContentTypeApplicationJSON {
		t.Errorf("Wrong Content-Type: got %v want %v", contentType, httpconsts.ContentTypeApplicationJSON)
	}
}

func TestRequestTooLarge(t *testing.T) {
	data := base64.StdEncoding.EncodeToString((make([]byte, maxRequestSize+1)))
	s := fmt.Sprintf("{\"event-type\":\"order.created\",\"event-type-version\":\"v1\",\"event-id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"event-time\":\"2012-11-01T22:08:41+00:00\",\"data\":\"%s\"}", data)
	req, err := http.NewRequest(http.MethodPost, shared.EventsV1Path, strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler := NewEventsHandler(maxRequestSize)
	handler.ServeHTTP(recorder, req)
	if status := recorder.Code; status != http.StatusRequestEntityTooLarge {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusRequestEntityTooLarge)
	}
	if contentType := recorder.Result().Header.Get("Content-Type"); contentType != httpconsts.ContentTypeApplicationJSON {
		t.Errorf("Wrong Content-Type: got %v want %v", contentType, httpconsts.ContentTypeApplicationJSON)
	}
}

// http client mock
type HTTPClientMock struct{}

func (c *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	response := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("{\"event-id\":\"cea54510-8631-47f0-934a-0571495c12d0\"}"))),
	}
	return response, nil
}

func TestPropagateTraceHeaders(t *testing.T) {
	// request to downstream services
	var downstreamReq *http.Request

	// mock the http request provider
	httpRequestProviderMock := func(method, url string, body io.Reader) (*http.Request, error) {
		downstreamReq, _ = http.NewRequest(method, url, body)
		return downstreamReq, nil
	}

	// mock the http client provider
	httpClientProviderMock := func() httptools.HTTPClient { return new(HTTPClientMock) }

	// init event sender with mocks
	bus.InitEventSender(httpClientProviderMock, httpRequestProviderMock)

	// reset event sender default http client provider and http request provider
	defer func() { bus.InitEventSender(httptools.DefaultHTTPClientProvider, httptools.DefaultHTTPRequestProvider) }()

	// init source config

	sourceID, targetURLV1, targetURLV2 := "", "http://kyma-domain/v1/events", "http://kyma-domain/v2/events"
	bus.Init(sourceID, targetURLV1, targetURLV2)

	// simulate request from outside of event-service
	event := "{\"event-type\":\"order.created\",\"event-type-version\":\"v1\",\"event-id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"event-time\":\"2012-11-01T22:08:41+00:00\",\"data\":\"{'key':'value'}\"}"
	req, err := http.NewRequest(http.MethodPost, shared.EventsV1Path, strings.NewReader(event))

	// simulate trace headers added by envoy sidecar
	traceHeaderKey, traceHeaderVal := "x-b3-traceid", "0887296564d75cda"
	req.Header.Add(traceHeaderKey, traceHeaderVal)

	// add none-trace headers
	nonTraceHeaderKey, nonTraceHeaderVal := "key", "value"
	req.Header.Add(nonTraceHeaderKey, nonTraceHeaderVal)

	if err != nil {
		t.Fatal(err)
	}

	if downstreamReq != nil {
		t.Fatal("http request should have not be initialized at this point")
	}

	recorder := httptest.NewRecorder()
	handler := NewEventsHandler(maxRequestSize)
	handler.ServeHTTP(recorder, req)

	// trace headers should be added to downstream request headers
	if downstreamReq.Header.Get(traceHeaderKey) != traceHeaderVal {
		t.Fatal("http request to events service is missing trace headers")
	}

	// none-trace headers should not be added to downstream request headers
	if downstreamReq.Header.Get(nonTraceHeaderKey) != "" {
		t.Fatal("should not propagate non-trace headers")
	}
}

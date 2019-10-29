package v2

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

	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/event-service/internal/httptools"
)

const (
	maxRequestSize = 64 * 1024
)

// Test that a valid event request on a initialized/working bus yields a valid response
func TestEventOk(t *testing.T) {

	// mock the http request provider
	httpRequestProviderMock := func(method, url string, body io.Reader) (*http.Request, error) {
		var err error
		downstreamReq, err := http.NewRequest(method, url, body)
		if err != nil {
			t.Logf("Error: %v", err)
			return nil, err
		}
		return downstreamReq, nil
	}
	// mock the http client provider
	httpClientProviderMock := func() httptools.HTTPClient { return new(HTTPClientMock) }
	// init event sender with mocks
	bus.InitEventSender(httpClientProviderMock, httpRequestProviderMock)
	// reset event sender default http client provider and http request provider
	defer func() { bus.InitEventSender(httptools.DefaultHTTPClientProvider, httptools.DefaultHTTPRequestProvider) }()

	s := &stupidEventMock{
		Typ:              "order.created",
		Eventtypeversion: "v1",
		Specversion:      "0.3",
		Id:               "31109198-4d69-4ae0-972d-76117f3748c8",
		Time:             "2012-11-01T22:08:41+00:00",
		Data:             "foo",
	}
	ss := eventMockToString(t, s)

	req, err := http.NewRequest(http.MethodPost, shared.EventsV2Path, strings.NewReader(ss))
	// TODO(nachtmaar): use sendAndReceiveError method
	// then the following lines are not necessary anymore
	req.Header.Add("Content-Type", "application/cloudevents+json")
	// init source config
	sourceID, targetURLV1, targetURLV2 := "some source ref", "http://kyma-domain/v1/events", "http://kyma-domain/v2/events"
	if err := bus.Init(sourceID, targetURLV1, targetURLV2); err != nil {
		t.Fatalf("unable to init bus")
	}

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

func TestEventNoGateway(t *testing.T) {
	//saved := handleEvent
	//defer func() { handleEvent = saved }()
	//
	//handleEvent = func(parameters *apiv2.PublishEventParametersV2, response *api.PublishEventResponses,
	//	traceHeaders *map[string]string, forwardHeaders *map[string][]string) (err error) {
	//	ok := api.PublishResponse{EventID: "responseEventId"}
	//	response.Ok = &ok
	//	return
	//}

	s := &stupidEventMock{
		Typ:              "order.created",
		Eventtypeversion: "v1",
		Specversion:      "0.3",
		Id:               "31109198-4d69-4ae0-972d-76117f3748c8",
		Time:             "2012-11-01T22:08:41+00:00",
		Data:             "foo",
	}
	ss := eventMockToString(t, s)

	req, err := http.NewRequest(http.MethodPost, shared.EventsV2Path, strings.NewReader(ss))
	// TODO(nachtmaar): use sendAndReceiveError method
	// then the following lines are not necessary anymore
	req.Header.Add("Content-Type", "application/cloudevents+json")
	// init source config
	sourceID, targetURLV1, targetURLV2 := "some source ref", "http://kyma-domain/v1/events", "http://kyma-domain/v2/events"
	if err := bus.Init(sourceID, targetURLV1, targetURLV2); err != nil {
		t.Fatalf("unable to init bus")
	}

	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler := NewEventsHandler(maxRequestSize)
	handler.ServeHTTP(recorder, req)
	if status := recorder.Code; status != http.StatusBadGateway {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusBadGateway)
	}
	if contentType := recorder.Result().Header.Get("Content-Type"); contentType != httpconsts.ContentTypeApplicationJSON {
		t.Errorf("Wrong Content-Type: got %v want %v", contentType, httpconsts.ContentTypeApplicationJSON)
	}
}

func TestRequestTooLarge(t *testing.T) {
	// init source config
	sourceID, targetURLV1, targetURLV2 := "some.dummy.source", "http://kyma-domain/v1/events", "http://kyma-domain/v2/events"
	if err := bus.Init(sourceID, targetURLV1, targetURLV2); err != nil {
		t.Fatalf("unable to init bus")
	}

	data := base64.StdEncoding.EncodeToString(make([]byte, maxRequestSize+1))
	s := fmt.Sprintf("{\"type\":\"order.created\",\"eventtypeversion\":\"v1\",\"id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"time\":\"2012-11-01T22:08:41+00:00\",\"data\":\"%s\"}", data)
	req, err := http.NewRequest(http.MethodPost, shared.EventsV2Path, strings.NewReader(s))
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
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("{\"event-id\":\"cea54510-8631-47f0-934a-0571495c12d0\",\"reason\":\"Message successfully published to the channel\",\"status\":\"published\"}"))),
	}
	return response, nil
}

func TestPropagateTraceHeaders(t *testing.T) {
	// request to downstream services
	var downstreamReq *http.Request

	// mock the http request provider
	httpRequestProviderMock := func(method, url string, body io.Reader) (*http.Request, error) {
		var err error
		downstreamReq, err = http.NewRequest(method, url, body)
		if err != nil {
			t.Logf("Error: %v", err)
			return nil, err
		}
		return downstreamReq, nil
	}

	// mock the http client provider
	httpClientProviderMock := func() httptools.HTTPClient { return new(HTTPClientMock) }

	// init event sender with mocks
	bus.InitEventSender(httpClientProviderMock, httpRequestProviderMock)

	// reset event sender default http client provider and http request provider
	defer func() { bus.InitEventSender(httptools.DefaultHTTPClientProvider, httptools.DefaultHTTPRequestProvider) }()

	// init source config
	sourceID, targetURLV1, targetURLV2 := "some.dummy.source", "http://kyma-domain/v1/events", "http://kyma-domain/v2/events"
	if err := bus.Init(sourceID, targetURLV1, targetURLV2); err != nil {
		t.Fatalf("unable to init bus")
	}

	// simulate request from outside of event-service
	s := &stupidEventMock{
		Typ:              "order.created",
		Specversion:      "0.3",
		Eventtypeversion: "v1",
		Id:               "31109198-4d69-4ae0-972d-76117f3748c8",
		Time:             "2012-11-01T22:08:41+00:00",
		Data:             "{'key':'value'}",
	}
	ss := eventMockToString(t, s)
	req, err := http.NewRequest(http.MethodPost, shared.EventsV2Path, strings.NewReader(ss))
	if err != nil {
		t.Fatal(err)
	}

	// simulate trace headers added by envoy sidecar
	traceHeaderKey, traceHeaderVal := "x-b3-traceid", "0887296564d75cda"
	req.Header.Add("content-type", "application/cloudevents+json")
	req.Header.Add(traceHeaderKey, traceHeaderVal)

	// add none-trace headers
	nonTraceHeaderKey, nonTraceHeaderVal := "key", "value"
	req.Header.Add(nonTraceHeaderKey, nonTraceHeaderVal)

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

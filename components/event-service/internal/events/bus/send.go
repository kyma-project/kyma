package bus

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	cloudevents "github.com/cloudevents/sdk-go"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/event-service/internal/httptools"
)

var (
	httpClientProvider  = httptools.DefaultHTTPClientProvider
	httpRequestProvider = httptools.DefaultHTTPRequestProvider
)

type v2Sender struct {
	Client cloudevents.Client
}

func NewSender() (*v2Sender, error) {
	s := &v2Sender{}
	options := []cloudeventshttp.Option{
		cloudevents.WithTarget(eventsTargetURLV2),
		cloudevents.WithStructuredEncoding(),
	}

	t, err := cloudevents.NewHTTPTransport(options...)

	if err != nil {
		return nil, err
	}

	c, err := cloudevents.NewClient(t,
		cloudevents.WithTimeNow(),
	)
	if err != nil {
		return nil, err
	}

	s.Client = c
	return s, nil
}

// InitEventSender initializes an internal httpClientProvider and httpRequestProvider
func InitEventSender(clientProvider httptools.HTTPClientProvider, requestProvider httptools.HTTPRequestProvider) {
	httpClientProvider = clientProvider
	httpRequestProvider = requestProvider
}

// SendEvent sends the incoming request to the Sender
func SendEvent(apiVersion string, req interface{}, traceHeaders *map[string]string,
	forwardHeaders *map[string][]string) (*api.SendEventResponse, error) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(req)
	httpReq, err := httpRequestProvider(http.MethodPost, "", body)
	if err != nil {
		return nil, err
	}

	var reqURL *url.URL

	switch apiVersion {
	case "v1":
		reqURL, err = url.ParseRequestURI(eventsTargetURLV1)
	case "v2":
		reqURL, err = url.ParseRequestURI(eventsTargetURLV2)
	}

	if err != nil {
		return nil, err
	}
	httpReq.URL = reqURL

	headers := make(http.Header)
	headers = *forwardHeaders
	headers.Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJSONWithCharset)
	httpReq.Header = headers
	httpReq.Header.Add(httpconsts.HeaderXForwardedFor, httpReq.Host)
	httpReq.Header.Del(httpconsts.HeaderConnection)

	addTraceHeaders(httpReq, traceHeaders)

	resp, err := httpClientProvider().Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	response := api.SendEventResponse{}

	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &api.PublishResponse{}

		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}

		response.Ok = result

	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		result := &api.Error{}

		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}

		response.Error = result
	}

	return &response, nil
}

// Send an event to event-bus as CloudEvents 1.0 in structured encoding
// Use the client from the cloudevents sdk for sending
func (s *v2Sender) SendEventV2(event cloudevents.Event, traceHeaders map[string]string) (*cloudevents.Event, error) {
	ctx := cloudevents.ContextWithEncoding(context.Background(), cloudevents.Structured)

	for key, value := range traceHeaders {
		ctx = cloudevents.ContextWithHeader(ctx, key, value)
	}

	// TODO(k15r): is this how you convert to v1?
	//event = cloudevents.Event{
	//	Context: event.Context.AsV1(),
	//	Data:    event.Data,
	//}
	_, resp, err := s.Client.Send(ctx, event)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func addTraceHeaders(httpReq *http.Request, traceHeaders *map[string]string) {
	if traceHeaders != nil {
		for key, value := range *traceHeaders {
			httpReq.Header.Add(key, value)
		}
	}
}

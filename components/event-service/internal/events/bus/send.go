package bus

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/event-service/internal/httptools"
)

var (
	httpClientProvider  = httptools.DefaultHTTPClientProvider
	httpRequestProvider = httptools.DefaultHTTPRequestProvider
)

// InitEventSender initializes an internal httpClientProvider and httpRequestProvider
func InitEventSender(clientProvider httptools.HTTPClientProvider, requestProvider httptools.HTTPRequestProvider) {
	httpClientProvider = clientProvider
	httpRequestProvider = requestProvider
}

// SendEvent sends the incoming request to the Sender
func SendEvent(req interface{}, traceHeaders *map[string]string, forwardHeaders *map[string][]string) (*api.SendEventResponse, error) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(req)
	httpReq, err := httpRequestProvider(http.MethodPost, "", body)
	if err != nil {
		return nil, err
	}

	if reqURL, err := url.ParseRequestURI(eventsTargetURLV2); err != nil {
		return nil, err
	} else {
		httpReq.URL = reqURL
	}

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

func addTraceHeaders(httpReq *http.Request, traceHeaders *map[string]string) {
	if traceHeaders != nil {
		for key, value := range *traceHeaders {
			httpReq.Header.Add(key, value)
		}
	}
}

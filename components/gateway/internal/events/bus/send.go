package bus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kyma-project/kyma/components/gateway/internal/events/api"
	"github.com/kyma-project/kyma/components/gateway/internal/httpconsts"
	"github.com/kyma-project/kyma/components/gateway/internal/httptools"
)

var (
	httpClientProvider  = httptools.DefaultHttpClientProvider
	httpRequestProvider = httptools.DefaultHttpRequestProvider
)

func InitEventSender(clientProvider httptools.HttpClientProvider, requestProvider httptools.HttpRequestProvider) {
	httpClientProvider = clientProvider
	httpRequestProvider = requestProvider
}

// SendEvent sends the incoming request to the Sender
func SendEvent(req *api.SendEventParameters, traceHeaders *map[string]string) (*api.SendEventResponse, error) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(req)
	httpReq, err := httpRequestProvider(http.MethodPost, "", body)
	if err != nil {
		return nil, err
	}

	headers := make(http.Header)
	headers.Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
	httpReq.Header = headers

	reqURL, err := url.ParseRequestURI(eventsTargetURL)
	if err != nil {
		fmt.Println(err)
	}

	httpReq.URL = reqURL
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

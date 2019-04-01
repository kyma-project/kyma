package testkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"fmt"
)

const (
	retryCount             = 3
	requestTimeout         = 15 * time.Second
	retryDelay             = 1 * time.Second
	modifyIdentifierFormat = "%s-%d"
)

type MetadataServiceClient interface {
	CreateService(t *testing.T, serviceDetails ServiceDetails) (int, *PostServiceResponse, error)
	UpdateService(t *testing.T, idToUpdate string, updatedServiceDetails ServiceDetails) (int, error)
	DeleteService(t *testing.T, idToDelete string) (int, error)
	GetService(t *testing.T, serviceId string) (int, *ServiceDetails, error)
	GetAllServices(t *testing.T) (int, []Service, error)
}

type metadataServiceClient struct {
	url        string
	httpClient http.Client
}

func NewMetadataServiceClient(url string) MetadataServiceClient {
	httpClient := http.Client{
		Timeout: requestTimeout,
	}

	return &metadataServiceClient{
		url:        url,
		httpClient: httpClient,
	}
}

func (client *metadataServiceClient) CreateService(t *testing.T, serviceDetails ServiceDetails) (int, *PostServiceResponse, error) {
	requestData := requestData{
		method: http.MethodPost,
		url:    client.url,
		data:   &serviceDetails,
	}

	postResponse, err := client.requestWithRetries(t, requestData, newServiceDetailsRetryRequest, statusNotServerError)
	if err != nil {
		t.Log(err)
		return -1, nil, err
	}

	logResponse(t, postResponse)

	postResponseData := PostServiceResponse{}
	err = json.NewDecoder(postResponse.Body).Decode(&postResponseData)
	if err != nil {
		return -1, nil, err
	}

	return postResponse.StatusCode, &postResponseData, nil
}

func (client *metadataServiceClient) UpdateService(t *testing.T, idToUpdate string, updatedServiceDetails ServiceDetails) (int, error) {
	requestData := requestData{
		method: http.MethodPut,
		url:    client.url + "/" + idToUpdate,
		data:   &updatedServiceDetails,
	}

	putResponse, err := client.requestWithRetries(t, requestData, newServiceDetailsRetryRequest, statusNotServerError)
	if err != nil {
		t.Log(err)
		return -1, err
	}

	logResponse(t, putResponse)

	return putResponse.StatusCode, nil
}

func (client *metadataServiceClient) DeleteService(t *testing.T, idToDelete string) (int, error) {
	requestData := requestData{
		method: http.MethodDelete,
		url:    client.url + "/" + idToDelete,
		data:   nil,
	}

	deleteResponse, err := client.requestWithRetries(t, requestData, newEmptyRetryRequest, statusNotServerError)
	if err != nil {
		t.Log(err)
		return -1, err
	}

	logResponse(t, deleteResponse)

	return deleteResponse.StatusCode, nil
}

func (client *metadataServiceClient) GetService(t *testing.T, serviceId string) (int, *ServiceDetails, error) {
	condition := getSpecsPredicate(t, true, true, true)
	return client.getService(t, serviceId, condition)
}

func (client *metadataServiceClient) getService(t *testing.T, serviceId string, condition Predicate) (int, *ServiceDetails, error) {
	requestData := requestData{
		method: http.MethodGet,
		url:    client.url + "/" + serviceId,
		data:   nil,
	}

	getResponse, err := client.requestWithRetries(t, requestData, newEmptyRetryRequest, condition)
	if err != nil {
		t.Log(err)
		return -1, nil, err
	}

	logResponse(t, getResponse)

	serviceDetails := ServiceDetails{}
	err = json.NewDecoder(getResponse.Body).Decode(&serviceDetails)
	if err != nil {
		return -1, nil, err
	}

	return getResponse.StatusCode, &serviceDetails, nil
}

func (client *metadataServiceClient) GetAllServices(t *testing.T) (int, []Service, error) {
	requestData := requestData{
		method: http.MethodGet,
		url:    client.url,
		data:   nil,
	}

	getAllResponse, err := client.requestWithRetries(t, requestData, newEmptyRetryRequest, statusNotServerError)
	if err != nil {
		t.Log(err)
		return -1, nil, err
	}

	logResponse(t, getAllResponse)

	var existingServices []Service
	err = json.NewDecoder(getAllResponse.Body).Decode(&existingServices)
	if err != nil {
		return -1, nil, err
	}

	return getAllResponse.StatusCode, existingServices, nil
}

type requestData struct {
	method string
	url    string
	data   *ServiceDetails
}

func newServiceDetailsRetryRequest(data requestData, retry int) (*http.Request, error) {
	if data.data.Identifier != "" {
		data.data.Identifier = fmt.Sprintf(modifyIdentifierFormat, data.data.Identifier, retry)
	}

	body, err := json.Marshal(data.data)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(data.method, data.url, bytes.NewReader(body))
}

func newEmptyRetryRequest(data requestData, retry int) (*http.Request, error) {
	return http.NewRequest(data.method, data.url, nil)
}

type CreateRequestFunc func(data requestData, retry int) (*http.Request, error)

func (client *metadataServiceClient) requestWithRetries(t *testing.T, data requestData, createRequest CreateRequestFunc, condition Predicate) (*http.Response, error) {
	var response *http.Response
	var err error

	for i := 0; i < retryCount; i++ {
		if response != nil {
			response.Body.Close()
		}

		request, reqErr := createRequest(data, i)
		if reqErr != nil {
			t.Log(reqErr)
			return nil, reqErr
		}
		response, err := client.httpClient.Do(request)

		if response != nil && condition(response, err) {
			return response, err
		}

		time.Sleep(retryDelay)
	}

	if response == nil {
		return nil, errors.New("nil response returned once executing request")
	}

	return response, err
}

type Predicate func(response *http.Response, err error) bool

func statusNotServerError(response *http.Response, err error) bool {
	return err == nil && response.StatusCode < 500
}

func getSpecsPredicate(t *testing.T, expectApiSpec bool, expectEventsSpec bool, expectDocumentation bool) Predicate {
	return func(response *http.Response, err error) bool {
		if response.StatusCode == http.StatusOK {
			serviceDetails := ServiceDetails{}
			err = json.NewDecoder(response.Body).Decode(&serviceDetails)
			if err != nil {
				return false
			}

			apiSpecMatch := true
			if expectApiSpec {
				apiSpecMatch = serviceDetails.Api != nil && serviceDetails.Api.Spec != nil
			}

			eventsSpecMatch := true
			if expectEventsSpec {
				eventsSpecMatch = serviceDetails.Events != nil && serviceDetails.Events.Spec != nil
			}

			documentationMatch := true
			if expectDocumentation {
				documentationMatch = serviceDetails.Documentation != nil
			}

			return apiSpecMatch && eventsSpecMatch && documentationMatch
		}

		return err == nil && response.StatusCode < 500
	}
}

func logResponse(t *testing.T, resp *http.Response) {
	dump, err := httputil.DumpResponse(resp, false)

	if err != nil {
		t.Logf("failed to dump response, %s", err)
	} else {
		t.Logf("\n--------------------------------\n%s\n--------------------------------", dump)
	}

	if resp != nil && resp.Body != nil {
		if err = resp.Body.Close(); err != nil {
			t.Logf("failed to close response body, %s", err)
		}
	}
}

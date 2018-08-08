/*
 *  © 2018 SAP SE or an SAP affiliate company.
 *  All rights reserved.
 *  Please see http://www.sap.com/corporate-en/legal/copyright/index.epx for additional trademark information and
 *  notices.
 */
package testkit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/gojektech/heimdall"
)

const (
	retryCount            = 3
	requestTimeout        = 3 * time.Second
	backoffInterval       = 1 * time.Second
	maximumJitterInterval = 1 * time.Second
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
	httpClient heimdall.Client
}

func NewMetadataServiceClient(url string) MetadataServiceClient {
	backoff := heimdall.NewConstantBackoff(backoffInterval, maximumJitterInterval)
	retrier := heimdall.NewRetrier(backoff)

	client := heimdall.NewHTTPClient(requestTimeout)
	client.SetRetrier(retrier)
	client.SetRetryCount(retryCount)

	return &metadataServiceClient{
		url:        url,
		httpClient: client,
	}
}

func (client *metadataServiceClient) CreateService(t *testing.T, serviceDetails ServiceDetails) (int, *PostServiceResponse, error) {
	postBody, err := json.Marshal(serviceDetails)
	if err != nil {
		return -1, nil, err
	}

	postRequest, err := http.NewRequest(http.MethodPost, client.url, bytes.NewBuffer(postBody))
	if err != nil {
		return -1, nil, err
	}
	postRequest.Header.Add("Content-Type", "application/json")

	postResponse, err := client.httpClient.Do(postRequest)
	if err != nil {
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
	putBody, err := json.Marshal(updatedServiceDetails)
	if err != nil {
		return -1, err
	}

	putRequest, err := http.NewRequest(http.MethodPut, client.url+"/"+idToUpdate, bytes.NewBuffer(putBody))
	if err != nil {
		return -1, err
	}
	putRequest.Header.Set("Content-Type", "application/json")

	putResponse, err := client.httpClient.Do(putRequest)
	if err != nil {
		return -1, err
	}
	logResponse(t, putResponse)

	return putResponse.StatusCode, nil
}

func (client *metadataServiceClient) DeleteService(t *testing.T, idToDelete string) (int, error) {
	deleteRequest, err := http.NewRequest(http.MethodDelete, client.url+"/"+idToDelete, nil)
	if err != nil {
		return -1, err
	}

	deleteResponse, err := client.httpClient.Do(deleteRequest)
	if err != nil {
		return -1, err
	}
	logResponse(t, deleteResponse)

	return deleteResponse.StatusCode, nil
}

func (client *metadataServiceClient) GetService(t *testing.T, serviceId string) (int, *ServiceDetails, error) {
	getRequest, err := http.NewRequest(http.MethodGet, client.url+"/"+serviceId, nil)
	if err != nil {
		return -1, nil, err
	}

	getResponse, err := client.httpClient.Do(getRequest)
	if err != nil {
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
	getAllRequest, err := http.NewRequest(http.MethodGet, client.url, nil)
	if err != nil {
		return -1, nil, err
	}

	getAllResponse, err := client.httpClient.Do(getAllRequest)
	if err != nil {
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

func logResponse(t *testing.T, resp *http.Response) {
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		t.Logf("failed to dump response, %s", err)
	} else {
		t.Logf("\n--------------------------------\n%s\n--------------------------------", dump)
	}
}

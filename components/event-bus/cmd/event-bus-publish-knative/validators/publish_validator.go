package validators

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	v2 "github.com/kyma-project/kyma/components/event-bus/api/publish/v2"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

const (
	requestBodyTooLargeErrorMessage = "http: request body too large"
)

// ValidateRequestV1 validates the http.Request and returns an api.Request instance and an api.Error.
func ValidateRequestV1(r *http.Request) (*api.Request, *api.Error) {
	// validate the http method
	if r.Method != http.MethodPost {
		log.Fatalf("request method not supported: %v", r.Method)
		return nil, api.ErrorResponseBadRequest()
	}

	// validate the request body for nil
	if r.Body == nil {
		log.Println("request body is nil")
		return nil, api.ErrorResponseBadRequest()
	}

	// validate read the request body
	body, err := ioutil.ReadAll(r.Body)
	_ = r.Body.Close()
	if err != nil {
		log.Printf("failed to read request body: %v", err)

		if err.Error() == requestBodyTooLargeErrorMessage {
			return nil, api.ErrorResponseRequestBodyTooLarge()
		}

		return nil, api.ErrorResponseInternalServer()
	}

	// validate parse the request body
	publishRequest := &api.Request{}
	err = json.Unmarshal(body, publishRequest)
	if err != nil {
		return nil, api.ErrorResponseBadPayload()
	}

	return publishRequest, nil
}

// ValidateRequestV2 validates the http.Request and returns an api.Request instance and an api.Error.
func ValidateRequestV2(r *http.Request) (*v2.EventRequestV2, *api.Error) {
	// validate the http method
	if r.Method != http.MethodPost {
		log.Printf("request method not supported: %v", r.Method)
		return nil, api.ErrorResponseBadRequest()
	}

	// validate the request body for nil
	if r.Body == nil {
		log.Println("request body is nil")
		return nil, api.ErrorResponseBadRequest()
	}

	// validate read the request body
	body, err := ioutil.ReadAll(r.Body)
	_ = r.Body.Close()
	if err != nil {
		log.Printf("failed to read request body: %v", err)

		if err.Error() == requestBodyTooLargeErrorMessage {
			return nil, api.ErrorResponseRequestBodyTooLarge()
		}

		return nil, api.ErrorResponseInternalServer()
	}

	// validate parse the request body
	publishRequest := &v2.EventRequestV2{}
	err = json.Unmarshal(body, publishRequest)
	if err != nil {
		return nil, api.ErrorResponseBadPayload()
	}

	return publishRequest, nil
}

// ValidateChannelNameLength validates the channel name length.
func ValidateChannelNameLength(channelName *string, length int) *api.Error {
	if len(*channelName) > length {
		return util.ErrorInvalidChannelNameLength(length)
	}
	return nil
}

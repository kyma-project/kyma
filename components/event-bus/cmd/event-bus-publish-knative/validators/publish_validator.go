package validators

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
)

func ValidateRequest(r *http.Request) (*api.PublishRequest, *api.Error) {
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
		return nil, api.ErrorResponseInternalServer()
	}

	// validate parse the request body
	publishRequest := &api.PublishRequest{}
	err = json.Unmarshal(body, publishRequest)
	if err != nil {
		return nil, api.ErrorResponseBadPayload()
	}

	return publishRequest, nil
}
